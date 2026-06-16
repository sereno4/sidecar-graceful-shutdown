package main

import (
"encoding/json"
"io"
"log"
"math/rand"
"net/http"
"sync"
"time"

shhttp "sidecar-graceful-shutdown/pkg/shared/http"
)

type IngestRequest struct {
Metrics []map[string]interface{} `json:"metrics"`
}

type Backend struct {
mu       sync.RWMutex
received int
failRate float64
delay    time.Duration
}

func main() {

backend := &Backend{
failRate: 0.0,
delay:    0,
}

server := shhttp.NewServer(
"0.0.0.0:8080",
)

server.Handle("/ingest", func(w http.ResponseWriter, r *http.Request) {

delay := backend.getDelay()

if delay > 0 {
time.Sleep(delay)
}


if backend.shouldFail() {

log.Println(
"Simulando falha do backend",
)

shhttp.JSONResponse(
w,
http.StatusServiceUnavailable,
shhttp.Response{
Status: "error",
Message: "backend indisponivel",
},
)

return
}


body, err := io.ReadAll(r.Body)

if err != nil {

shhttp.JSONResponse(
w,
http.StatusBadRequest,
shhttp.Response{
Status: "bad request",
},
)

return
}


var req IngestRequest

if err := json.Unmarshal(body, &req); err != nil {

log.Printf(
"Payload recebido: %s",
string(body),
)
}


backend.mu.Lock()

backend.received++

count := backend.received

backend.mu.Unlock()


log.Printf(
"Métricas recebidas: total=%d",
count,
)


shhttp.JSONResponse(
w,
http.StatusOK,
map[string]interface{}{
"status": "ok",
"received": count,
},
)
})


server.Handle("/received-count", func(w http.ResponseWriter, r *http.Request) {

backend.mu.RLock()

count := backend.received

backend.mu.RUnlock()


shhttp.JSONResponse(
w,
http.StatusOK,
map[string]interface{}{
"received": count,
},
)
})


server.Handle("/healthz", func(w http.ResponseWriter, r *http.Request) {

shhttp.JSONResponse(
w,
http.StatusOK,
shhttp.HealthResponse{
Status: "healthy",
Timestamp: time.Now().Unix(),
},
)
})


server.Handle("/chaos", func(w http.ResponseWriter, r *http.Request) {

if r.Method != http.MethodPost {

shhttp.JSONResponse(
w,
http.StatusMethodNotAllowed,
shhttp.Response{
Status: "method not allowed",
},
)

return
}


var params struct {

FailRate float64 `json:"fail_rate"`

DelayMs int `json:"delay_ms"`
}


body, _ := io.ReadAll(r.Body)

json.Unmarshal(
body,
&params,
)


backend.mu.Lock()

backend.failRate = params.FailRate

backend.delay =
time.Duration(params.DelayMs) *
time.Millisecond

backend.mu.Unlock()


log.Printf(
"Chaos ativado failRate=%.2f delay=%v",
params.FailRate,
time.Duration(params.DelayMs)*time.Millisecond,
)


shhttp.JSONResponse(
w,
http.StatusOK,
map[string]interface{}{
"status": "chaos activated",
},
)
})


server.Start()


log.Println(
"Mock Telemetry Backend rodando :8080",
)


select {}
}


func (b *Backend) shouldFail() bool {

b.mu.RLock()

rate := b.failRate

b.mu.RUnlock()


return rand.Float64() < rate
}


func (b *Backend) getDelay() time.Duration {

b.mu.RLock()

defer b.mu.RUnlock()

return b.delay
}
