package http

import (
"context"
"net/http"
"testing"
"time"
)

func TestServer_StartShutdown(t *testing.T) {
srv := NewServer("127.0.0.1:18080")

srv.Handle("/test", func(w http.ResponseWriter, r *http.Request) {
JSONResponse(w, http.StatusOK, Response{
Status: "ok",
})
})

srv.Start()
time.Sleep(100 * time.Millisecond)

resp, err := http.Get("http://127.0.0.1:18080/test")
if err != nil {
t.Fatalf("request falhou: %v", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
t.Errorf("status esperado %d, recebeu %d",
http.StatusOK,
resp.StatusCode,
)
}

if err := srv.Shutdown(2 * time.Second); err != nil {
t.Errorf("shutdown falhou: %v", err)
}
}

func TestClient_CheckHealth(t *testing.T) {
srv := NewServer("127.0.0.1:18081")

srv.Handle("/healthz", func(w http.ResponseWriter, r *http.Request) {
JSONResponse(w, http.StatusOK, HealthResponse{
Status:    "healthy",
Timestamp: time.Now().Unix(),
})
})

srv.Start()
time.Sleep(100 * time.Millisecond)

client := NewClient("http://127.0.0.1:18081")

ctx, cancel := context.WithTimeout(
context.Background(),
2*time.Second,
)
defer cancel()

if !client.CheckHealth(ctx) {
t.Error("health check deveria retornar true")
}

if err := srv.Shutdown(1 * time.Second); err != nil {
t.Errorf("shutdown falhou: %v", err)
}
}

func TestNewClient(t *testing.T) {
client := NewClient("http://localhost:8080")

if client == nil {
t.Fatal("client não deveria ser nil")
}

if client.baseURL != "http://localhost:8080" {
t.Errorf(
"baseURL esperada http://localhost:8080, recebeu %s",
client.baseURL,
)
}
}
