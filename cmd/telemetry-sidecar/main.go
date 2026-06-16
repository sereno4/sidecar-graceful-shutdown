package main

import (
"context"
"fmt"
"log"
"net/http"
"os"
"os/signal"
"sidecar-graceful-shutdown/pkg/config"
"sidecar-graceful-shutdown/pkg/shared/atomicfile"
"sidecar-graceful-shutdown/pkg/shared/graceperiod"
shhttp "sidecar-graceful-shutdown/pkg/shared/http"
"sync"
"syscall"
"time"
)

type Metric struct {
Timestamp int64             `json:"timestamp"`
Name      string            `json:"name"`
Value     float64           `json:"value"`
Labels    map[string]string `json:"labels,omitempty"`
}

type Sidecar struct {
cfg         *config.Config
buffer      []Metric
bufferMu    sync.Mutex
graceMgr    *graceperiod.Manager
batchDone   bool
batchDoneMu sync.Mutex
server      *shhttp.Server
stopCollect chan bool
shutdown    chan bool
}

func main() {
cfg, err := config.Load()
if err != nil {
log.Fatalf("carregar config: %v", err)
}

log.Printf("Telemetry Sidecar iniciando: buffer=%d backend=%s", cfg.BufferSize, cfg.BackendURL)

sidecar := &Sidecar{
cfg:         cfg,
buffer:      make([]Metric, 0, cfg.BufferSize),
graceMgr:    graceperiod.New(cfg.GracePeriodSeconds, 5*time.Second),
stopCollect: make(chan bool, 1),
shutdown:    make(chan bool, 1),
}

server := shhttp.NewServer(fmt.Sprintf("0.0.0.0:%d", cfg.SidecarPort))

server.Handle("/metrics", func(w http.ResponseWriter, r *http.Request) {
shhttp.JSONResponse(w, 200, map[string]interface{}{
"buffer_size": len(sidecar.buffer),
"batch_done":  sidecar.isBatchDone(),
})
})

server.Handle("/batch-done", func(w http.ResponseWriter, r *http.Request) {
log.Println(">>> Batch notificou que terminou!")
sidecar.setBatchDone(true)
sidecar.stopCollect <- true
shhttp.JSONResponse(w, 200, shhttp.Response{Status: "acknowledged"})
})

// ⭐ NOVO: /shutdown → sidecar sai
server.Handle("/shutdown", func(w http.ResponseWriter, r *http.Request) {
log.Println(">>> Batch pediu para sair. Finalizando...")
sidecar.shutdown <- true
shhttp.JSONResponse(w, 200, shhttp.Response{Status: "shutting down"})
})

server.Handle("/healthz", func(w http.ResponseWriter, r *http.Request) {
shhttp.JSONResponse(w, 200, shhttp.HealthResponse{
Status:    "healthy",
Timestamp: time.Now().Unix(),
})
})

server.Start()

// Inicia coleta
go sidecar.simulateCollection()

// ⭐ AGUARDA /shutdown ou SIGTERM
log.Println("Sidecar aguardando evento de shutdown...")
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

select {
case <-sidecar.shutdown:
log.Println(">>> Shutdown recebido via HTTP. Saindo...")
case sig := <-sigChan:
log.Printf(">>> Sinal recebido: %v. Iniciando graceful shutdown...", sig)
sidecar.graceMgr.OnSIGTERM()
sidecar.handleShutdown()
}

server.Shutdown(2 * time.Second)
log.Println("Sidecar finalizado com sucesso.")
}

func (s *Sidecar) simulateCollection() {
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

itemCount := 0
for {
select {
case <-s.stopCollect:
log.Println(">>> Coleta interrompida (batch done).")
return
case <-ticker.C:
if s.isBatchDone() {
log.Println(">>> Batch done detectado. Parando coleta.")
return
}

s.bufferMu.Lock()
s.buffer = append(s.buffer, Metric{
Timestamp: time.Now().Unix(),
Name:      "batch_item_processed",
Value:     float64(itemCount),
Labels:    map[string]string{"job": "batch-test"},
})
s.bufferMu.Unlock()

itemCount++
log.Printf("Metrica coletada: item %d", itemCount)
}
}
}

func (s *Sidecar) handleShutdown() {
log.Printf("Grace period manager: %s", s.graceMgr.String())

if !s.isBatchDone() {
log.Println("Aguardando batch done antes do flush...")
s.waitForBatchDone()
}

ctx, cancel := s.graceMgr.FlushContext()
defer cancel()

if err := s.flushBuffer(ctx); err != nil {
log.Printf("Flush final falhou: %v", err)
s.saveToFallback()
}

s.notifyBatchFlushed()
s.createFlushedFallback()
}

func (s *Sidecar) waitForBatchDone() {
batchURL := fmt.Sprintf("http://localhost:%d", s.cfg.ServerPort)
client := shhttp.NewClient(batchURL)

for !s.graceMgr.IsExpired() {
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
healthy := client.CheckHealth(ctx)
cancel()

if !healthy {
if s.checkBatchDoneFallback() {
log.Println("Batch done detectado via emptyDir fallback")
s.setBatchDone(true)
return
}
}

if s.isBatchDone() {
return
}

time.Sleep(500 * time.Millisecond)
}
}

func (s *Sidecar) checkBatchDoneFallback() bool {
doneFile := fmt.Sprintf("%s/.batch-done", s.cfg.SharedVolumePath)
return atomicfile.Exists(doneFile)
}

func (s *Sidecar) flushBuffer(ctx context.Context) error {
s.bufferMu.Lock()
defer s.bufferMu.Unlock()

if len(s.buffer) == 0 {
log.Println("Buffer vazio. Nada para flushar.")
return nil
}

log.Printf(">>> Flushando %d metricas...", len(s.buffer))

select {
case <-ctx.Done():
return fmt.Errorf("contexto cancelado durante flush")
default:
}

time.Sleep(100 * time.Millisecond)

log.Printf(">>> Flush completo: %d metricas enviadas", len(s.buffer))
s.buffer = s.buffer[:0]
return nil
}

func (s *Sidecar) saveToFallback() {
s.bufferMu.Lock()
defer s.bufferMu.Unlock()

if len(s.buffer) == 0 {
return
}

fallbackDir := "/data/fallback"
os.MkdirAll(fallbackDir, 0755)

filename := fmt.Sprintf("%s/metrics-%d.json", fallbackDir, time.Now().Unix())
data := fmt.Sprintf(`{"count": %d, "timestamp": %d}`, len(s.buffer), time.Now().Unix())

if err := atomicfile.WriteAtomically(filename, []byte(data)); err != nil {
log.Printf("ERRO: falha ao salvar fallback: %v", err)
return
}

log.Printf("Metricas salvas em fallback: %s", filename)
}

func (s *Sidecar) notifyBatchFlushed() {
batchURL := fmt.Sprintf("http://localhost:%d", s.cfg.ServerPort)
client := shhttp.NewClient(batchURL)

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

resp, err := client.Post(ctx, "/flushed", shhttp.FlushedResponse{
Flushed:   true,
Timestamp: time.Now().Unix(),
})
if err != nil {
log.Printf("Falha ao notificar batch: %v", err)
return
}
defer resp.Body.Close()

log.Println(">>> Batch notificado que flush completou")
}

func (s *Sidecar) createFlushedFallback() {
flushedFile := fmt.Sprintf("%s/.sidecar-flushed", s.cfg.SharedVolumePath)
timestamp := fmt.Sprintf("%d", time.Now().Unix())

if err := atomicfile.WriteAtomically(flushedFile, []byte(timestamp)); err != nil {
log.Printf("ERRO: falha ao criar .sidecar-flushed: %v", err)
}
}

func (s *Sidecar) isBatchDone() bool {
s.batchDoneMu.Lock()
defer s.batchDoneMu.Unlock()
return s.batchDone
}

func (s *Sidecar) setBatchDone(done bool) {
s.batchDoneMu.Lock()
defer s.batchDoneMu.Unlock()
s.batchDone = done
}
