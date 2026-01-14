package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Runtime provides HTTP execution capabilities for the CLI
type Runtime struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
	headersMu  sync.RWMutex
	Timeout    time.Duration
	Output     io.Writer
}

// New creates a new Runtime with the given configuration
func New(baseURL string, timeout time.Duration) *Runtime {
	return &Runtime{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Headers: make(map[string]string),
		Timeout: timeout,
		Output:  os.Stdout,
	}
}

// AddHeader adds a header to all requests
func (r *Runtime) AddHeader(key, value string) {
	r.headersMu.Lock()
	defer r.headersMu.Unlock()
	r.Headers[key] = value
}

// Do executes an HTTP request and handles the response
func (r *Runtime) Do(ctx context.Context, req *Request) error {
	httpReq, err := req.Build(ctx, r.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Add runtime headers
	r.headersMu.RLock()
	for k, v := range r.Headers {
		httpReq.Header.Set(k, v)
	}
	r.headersMu.RUnlock()

	resp, err := r.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for SSE response
	contentType := resp.Header.Get("Content-Type")
	if isEventStream(contentType) {
		return handleSSE(resp.Body, r.Output)
	}

	// Handle regular response
	return handleResponse(resp, r.Output)
}
