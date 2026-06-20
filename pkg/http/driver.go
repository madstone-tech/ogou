package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/madstone-tech/ogou/internal/template"
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

	// Interpolate path
	path, err := template.Interpolate(step.Path, sc)
	if err != nil {
		res.EndTime = time.Now()
		res.Latency = res.EndTime.Sub(res.StartTime)
		res.Error = fmt.Sprintf("template path: %v", err)
		res.Success = false
		return res, nil
	}
	url := strings.TrimSuffix(baseURL, "/") + path

	// Resolve body
	var bodyBytes []byte
	if step.Body != nil {
		bodyStr, err := d.resolveBody(step.Body, sc)
		if err != nil {
			res.EndTime = time.Now()
			res.Latency = res.EndTime.Sub(res.StartTime)
			res.Error = fmt.Sprintf("template body: %v", err)
			res.Success = false
			return res, nil
		}
		bodyBytes = []byte(bodyStr)
	}

	method := step.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		res.EndTime = time.Now()
		res.Latency = res.EndTime.Sub(res.StartTime)
		res.Error = fmt.Sprintf("build request: %v", err)
		res.Success = false
		return res, nil
	}

	// Interpolate and set headers
	for k, v := range step.Headers {
		interpolatedVal, err := template.Interpolate(v, sc)
		if err != nil {
			res.EndTime = time.Now()
			res.Latency = res.EndTime.Sub(res.StartTime)
			res.Error = fmt.Sprintf("template header %q: %v", k, err)
			res.Success = false
			return res, nil
		}
		req.Header.Set(k, interpolatedVal)
	}

	resp, err := d.client.Do(req)
	res.EndTime = time.Now()
	res.Latency = res.EndTime.Sub(res.StartTime)

	if err != nil {
		res.Error = fmt.Sprintf("request failed: %v", err)
		res.Success = false
		return res, nil
	}
	defer func() { _ = resp.Body.Close() }()

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

func (d *Driver) resolveBody(bs *engine.BodySource, sc *engine.StepContext) (string, error) {
	raw := ""
	switch {
	case bs.File != "":
		b, err := os.ReadFile(bs.File)
		if err != nil {
			return "", err
		}
		raw = string(b)
	case bs.Template != "":
		raw = bs.Template
	case bs.Inline != "":
		raw = bs.Inline
	}
	interpolated, err := template.Interpolate(raw, sc)
	if err != nil {
		return "", err
	}
	return interpolated, nil
}

// Close is a no-op for HTTP; the client is reused.
func (d *Driver) Close() error {
	return nil
}
