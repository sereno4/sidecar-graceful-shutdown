package http

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"time"
)

// Client HTTP com timeout
type Client struct {
client  *http.Client
baseURL string
}

// NewClient cria client para endpoint
func NewClient(baseURL string) *Client {
return &Client{
client: &http.Client{
Timeout: 5 * time.Second,
},
baseURL: baseURL,
}
}

// Post envia POST JSON
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
url := c.baseURL + path

data, err := json.Marshal(body)
if err != nil {
return nil, fmt.Errorf("marshal json: %w", err)
}

req, err := http.NewRequestWithContext(
ctx,
http.MethodPost,
url,
bytes.NewReader(data),
)
if err != nil {
return nil, fmt.Errorf("criar request: %w", err)
}

req.Header.Set("Content-Type", "application/json")

return c.client.Do(req)
}

// Get envia GET
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
url := c.baseURL + path

req, err := http.NewRequestWithContext(
ctx,
http.MethodGet,
url,
nil,
)
if err != nil {
return nil, fmt.Errorf("criar request: %w", err)
}

return c.client.Do(req)
}

// CheckHealth verifica se endpoint está saudável
func (c *Client) CheckHealth(ctx context.Context) bool {
resp, err := c.Get(ctx, "/healthz")
if err != nil {
return false
}
defer resp.Body.Close()

return resp.StatusCode == http.StatusOK
}

// ReadBody lê e fecha body
func ReadBody(resp *http.Response) ([]byte, error) {
defer resp.Body.Close()

return io.ReadAll(resp.Body)
}
