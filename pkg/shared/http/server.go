package http

import (
"context"
"encoding/json"
"log"
"net/http"
"time"
)

// Server HTTP leve com graceful shutdown
type Server struct {
server *http.Server
mux    *http.ServeMux
}

// NewServer cria servidor com timeout padrão
func NewServer(addr string) *Server {
mux := http.NewServeMux()

return &Server{
mux: mux,
server: &http.Server{
Addr:         addr,
Handler:      mux,
ReadTimeout:  5 * time.Second,
WriteTimeout: 10 * time.Second,
IdleTimeout:  120 * time.Second,
},
}
}

// Handle registra rota
func (s *Server) Handle(pattern string, handler http.HandlerFunc) {
s.mux.HandleFunc(pattern, handler)
}

// Start inicia servidor em goroutine
func (s *Server) Start() {
go func() {
log.Printf("HTTP server iniciando em %s", s.server.Addr)

if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
log.Printf("HTTP server erro: %v", err)
}
}()
}

// Shutdown graceful com timeout
func (s *Server) Shutdown(timeout time.Duration) error {
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

return s.server.Shutdown(ctx)
}

// Response padrão para APIs
type Response struct {
Status  string `json:"status"`
Message string `json:"message,omitempty"`
}

// JSONResponse escreve resposta JSON
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)

if err := json.NewEncoder(w).Encode(data); err != nil {
log.Printf("erro ao serializar JSON: %v", err)
}
}

// HealthResponse para /healthz
type HealthResponse struct {
Status    string `json:"status"`
Timestamp int64  `json:"timestamp"`
}

// DoneRequest sinalização de batch concluído
type DoneRequest struct {
Timestamp int64 `json:"timestamp"`
Items     int   `json:"items_processed"`
}

// FlushedResponse confirmação de flush
type FlushedResponse struct {
Flushed   bool  `json:"flushed"`
Timestamp int64 `json:"timestamp"`
Metrics   int   `json:"metrics_count"`
}
