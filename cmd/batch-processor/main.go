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
shhttp "sidecar-graceful-shutdown/pkg/shared/http"
"syscall"
"time"
)

func main() {
cfg, err := config.Load()
if err != nil {
log.Fatalf("carregar config: %v", err)
}

log.Printf("Batch Processor iniciando: duration=%v items=%d", cfg.BatchDuration, cfg.BatchItems)

done := make(chan bool, 1)
flushed := make(chan bool, 1)

server := shhttp.NewServer(fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort))

server.Handle("/healthz", func(w http.ResponseWriter, r *http.Request) {
shhttp.JSONResponse(w, 200, shhttp.HealthResponse{
Status:    "healthy",
Timestamp: time.Now().Unix(),
})
})

server.Handle("/done", func(w http.ResponseWriter, r *http.Request) {
log.Println("Sidecar confirmou recebimento de .batch-done")
shhttp.JSONResponse(w, 200, shhttp.Response{Status: "acknowledged"})
})

server.Handle("/flushed", func(w http.ResponseWriter, r *http.Request) {
log.Println(">>> Sidecar confirmou flush completo!")
flushed <- true
shhttp.JSONResponse(w, 200, shhttp.FlushedResponse{
Flushed:   true,
Timestamp: time.Now().Unix(),
})
})

server.Start()

go processBatch(cfg, done)

// Aguarda SIGTERM ou processamento natural
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

select {
case sig := <-sigChan:
log.Printf("SIGTERM/SIGINT recebido (%v)!", sig)
case <-done:
log.Println("Processamento completo. Notificando sidecar...")
}

// Notifica sidecar que batch terminou
if err := notifySidecar(cfg, done); err != nil {
log.Printf("Falha ao notificar sidecar via HTTP: %v", err)
log.Println("Usando fallback emptyDir...")
notifySidecarFallback(cfg)
}

// Aguarda /flushed do sidecar
log.Println("Aguardando sidecar confirmar flush...")
select {
case <-flushed:
log.Println(">>> Sidecar confirmou flush!")
case <-time.After(30 * time.Second):
log.Println("Timeout aguardando sidecar.")
}

// ⭐ NOVO: Pede ao sidecar para sair
log.Println("Pedindo ao sidecar para sair...")
if err := shutdownSidecar(cfg); err != nil {
log.Printf("Falha ao pedir shutdown do sidecar: %v", err)
}

server.Shutdown(2 * time.Second)
log.Println("Batch Processor finalizado.")
}

func processBatch(cfg *config.Config, done chan<- bool) {
log.Printf("Iniciando processamento de %d items...", cfg.BatchItems)
time.Sleep(cfg.BatchDuration)
log.Printf("Processamento completo: %d items processados", cfg.BatchItems)
done <- true
}

func notifySidecar(cfg *config.Config, done <-chan bool) error {
select {
case <-done:
default:
return fmt.Errorf("processamento ainda nao terminou")
}

sidecarURL := fmt.Sprintf("http://localhost:%d", cfg.SidecarPort)
client := shhttp.NewClient(sidecarURL)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req := shhttp.DoneRequest{
Timestamp: time.Now().Unix(),
Items:     cfg.BatchItems,
}

resp, err := client.Post(ctx, "/batch-done", req)
if err != nil {
return fmt.Errorf("POST /batch-done: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != 200 {
return fmt.Errorf("sidecar retornou %d", resp.StatusCode)
}

log.Println("Sidecar notificado com sucesso via HTTP")
return nil
}

func shutdownSidecar(cfg *config.Config) error {
sidecarURL := fmt.Sprintf("http://localhost:%d", cfg.SidecarPort)
client := shhttp.NewClient(sidecarURL)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.Post(ctx, "/shutdown", shhttp.Response{Status: "shutdown"})
if err != nil {
return fmt.Errorf("POST /shutdown: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != 200 {
return fmt.Errorf("sidecar retornou %d", resp.StatusCode)
}

log.Println("Sidecar pediu para sair com sucesso")
return nil
}

func notifySidecarFallback(cfg *config.Config) {
doneFile := fmt.Sprintf("%s/.batch-done", cfg.SharedVolumePath)
timestamp := fmt.Sprintf("%d", time.Now().Unix())

if err := atomicfile.WriteAtomically(doneFile, []byte(timestamp)); err != nil {
log.Printf("ERRO CRITICO: falha ao escrever fallback: %v", err)
return
}

log.Println("Fallback .batch-done escrito com sucesso")
}
