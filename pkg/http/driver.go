package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
)

// Driver implements engine.Driver for HTTP/1.1 and HTTP/2.
type Driver struct {
	client *http.Client
}

// NewDriver creates an HTTP driver with the given timeout.
func NewDriver(timeout time.Duration) *Driver {
	return &Driver{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the driver protocol name.
func (d *Driver) Name() string {
	return "http"
}

// Execute runs an HTTP step and returns an engine result.
func (d *Driver) Execute(ctx context.Context, baseURL string, step engine.Step, sc *engine.StepContext) (*engine.Result, error) {
	res := &engine.Result{
		StepName:     step.Name,
		StartTime:    time.Now(),
		CapturedVars: make(map[string]any),
	}

	url := strings.TrimSuffix(baseURL, "/") + step.Path
	var bodyBytes []byte
	if step.Body != nil {
		bodyBytes, _ = step.Body.Body()
	}
	req, err := http.NewRequestWithContext(ctx, step.Method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		res.EndTime = time.Now()
		res.Latency = res.EndTime.Sub(res.StartTime)
		res.Error = fmt.Sprintf("build request: %v", err)
		return res, nil
	}

	for k, v := range step.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	res.EndTime = time.Now()
	res.Latency = res.EndTime.Sub(res.StartTime)

	if err != nil {
		res.Error = fmt.Sprintf("request failed: %v", err)
		return res, nil
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	res.StatusCode = resp.StatusCode

	// Simple assertions
	res.Success = true
	for _, a := range step.Assertions {
		if a.Status != nil && resp.StatusCode != *a.Status {
			res.Success = false
			res.Error = fmt.Sprintf("status %d, want %d", resp.StatusCode, *a.Status)
		}
	}

	return res, nil
}

// Close is a no-op for HTTP; the client is reused.
func (d *Driver) Close() error {
	return nil
}
