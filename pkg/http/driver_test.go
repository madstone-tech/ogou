package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
)

func TestDriverExecute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "health",
		Method: "GET",
		Path:   "/",
		Assertions: []engine.Assertion{
			{Status: intPtr(200)},
		},
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
	if res.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
}

func TestDriverAssertionFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "fail",
		Method: "GET",
		Path:   "/",
		Assertions: []engine.Assertion{
			{Status: intPtr(200)},
		},
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure due to status mismatch")
	}
	if res.StatusCode != 500 {
		t.Fatalf("expected status 500, got %d", res.StatusCode)
	}
}

func TestDriverNilBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "nil-body",
		Method: "GET",
		Path:   "/",
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
}

func TestDriverMethods(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, m := range methods {
		step := engine.Step{
			Name:   m,
			Method: m,
			Path:   "/",
		}
		res, err := driver.Execute(context.Background(), ts.URL, step, sc)
		if err != nil {
			t.Fatalf("method %s: unexpected error: %v", m, err)
		}
		if !res.Success {
			t.Fatalf("method %s: expected success, got error: %s", m, res.Error)
		}
	}
}

func TestDriverTemplateInterpolation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	sc.VUID = 42
	sc.Iteration = 3

	step := engine.Step{
		Name:   "template",
		Method: "GET",
		Path:   "/users/{{.VU.ID}}/orders/{{.VU.Iteration}}",
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
}

func intPtr(i int) *int {
	return &i
}

func TestDriverMethodDefault(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "default-method",
		Method: "",
		Path:   "/",
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
}

func TestDriverBodyTemplate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	sc.VUID = 7
	step := engine.Step{
		Name:   "body-template",
		Method: "POST",
		Path:   "/",
		Body:   &engine.BodySource{Inline: `{"id":{{.VU.ID}}}`},
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
}

func TestDriverBadPathTemplate(t *testing.T) {
	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "bad-template",
		Method: "GET",
		Path:   "/{{.Bad}}",
	}

	res, err := driver.Execute(context.Background(), "http://localhost", step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure for bad template")
	}
}

func TestDriverName(t *testing.T) {
	d := NewDriver(10 * time.Second)
	if d.Name() != "http" {
		t.Fatalf("expected 'http', got %s", d.Name())
	}
}

func TestDriverClose(t *testing.T) {
	d := NewDriver(10 * time.Second)
	if err := d.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDriverBodyFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "htload-body-*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	if _, err := f.WriteString(`{"hello":"world"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = f.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "file-body",
		Method: "POST",
		Path:   "/",
		Body:   &engine.BodySource{File: f.Name()},
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
}

func TestDriverBodyTemplateInterpolationError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "bad-body-template",
		Method: "POST",
		Path:   "/",
		Body:   &engine.BodySource{Inline: `{{.NonExistent.Field}}`},
	}

	bodyStr, bodyErr := driver.resolveBody(step.Body, sc)
	t.Logf("bodyErr=%v bodyStr=%q", bodyErr, bodyStr)

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	_ = err
	t.Logf("res.Success=%v res.Error=%q", res.Success, res.Error)
	if res.Success {
		t.Fatal("expected failure for bad body template")
	}
}

func TestDriverBadHeaderTemplate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:    "bad-header",
		Method:  "GET",
		Path:    "/",
		Headers: map[string]string{"X-Bad": "{{.NoSuchField}}"},
	}

	res, err := driver.Execute(context.Background(), ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure for bad header template")
	}
}

func TestDriverRequestCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "canceled",
		Method: "GET",
		Path:   "/",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := driver.Execute(ctx, ts.URL, step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure for canceled context")
	}
}

func TestDriverInvalidMethod(t *testing.T) {
	driver := NewDriver(10 * time.Second)
	sc := engine.NewStepContext()
	step := engine.Step{
		Name:   "invalid-method",
		Method: "GET\nPOST",
		Path:   "/",
	}

	res, err := driver.Execute(context.Background(), "http://localhost", step, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure for invalid method")
	}
}
