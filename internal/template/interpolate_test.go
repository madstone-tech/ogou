package template

import (
	"strings"
	"testing"

	"github.com/madstone-tech/ogou/pkg/engine"
)

func TestInterpolateDistinctValues(t *testing.T) {
	sc := engine.NewStepContext()
	sc.VUID = 1
	sc.Iteration = 2
	sc.Vars["orgID"] = "org-123"

	tmpl := `{"user":"user-{{.VU.ID}}","iter":{{.VU.Iteration}},"org":"{{.Vars.orgID}}"}`
	result, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `"user":"user-1"`) {
		t.Fatalf("expected VU.ID=1, got: %s", result)
	}
	if !strings.Contains(result, `"iter":2`) {
		t.Fatalf("expected VU.Iteration=2, got: %s", result)
	}
	if !strings.Contains(result, `"org":"org-123"`) {
		t.Fatalf("expected Vars.orgID=org-123, got: %s", result)
	}
}

func TestInterpolateUUID(t *testing.T) {
	sc := engine.NewStepContext()
	tmpl := `{"id":"{{UUID}}"}`
	result1, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result2, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1 == result2 {
		t.Fatal("expected distinct UUIDs")
	}
}

func TestInterpolateRandomString(t *testing.T) {
	sc := engine.NewStepContext()
	tmpl := `{{RandomString 8}}`
	result1, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result2, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1 == result2 {
		t.Fatal("expected distinct random strings")
	}
	if len(result1) != 8 {
		t.Fatalf("expected length 8, got %d", len(result1))
	}
}

func TestInterpolateError(t *testing.T) {
	sc := engine.NewStepContext()
	tmpl := `{{InvalidFunc}}`
	_, err := Interpolate(tmpl, sc)
	if err == nil {
		t.Fatal("expected error for invalid template function")
	}
}

func TestInterpolateEmpty(t *testing.T) {
	sc := engine.NewStepContext()
	result, err := Interpolate("", sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string, got: %s", result)
	}
}

func TestInterpolateNoMarkers(t *testing.T) {
	sc := engine.NewStepContext()
	tmpl := `plain text without templates`
	result, err := Interpolate(tmpl, sc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != tmpl {
		t.Fatalf("expected passthrough, got: %s", result)
	}
}
