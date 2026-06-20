// Package template provides text/template interpolation for htload scenario values.
package template

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/madstone-tech/ogou/pkg/engine"
)

// Interpolate evaluates a template string using values from the StepContext and VU state.
func Interpolate(input string, sc *engine.StepContext) (string, error) {
	if input == "" {
		return "", nil
	}

	// Fast path: no template markers
	if !containsTemplateMarkers(input) {
		return input, nil
	}

	tmpl, err := template.New("step").Funcs(funcMap()).Parse(input)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := templateData{
		VU: templateVU{
			ID:        sc.VUID,
			Iteration: sc.Iteration,
		},
		Vars: copyMap(sc.Vars),
		Env:  envMap,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func containsTemplateMarkers(s string) bool {
	return bytes.Contains([]byte(s), []byte("{{"))
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"UUID":          uuid.NewString,
		"RandomString":  randomString,
		"Timestamp":     func() int64 { return time.Now().Unix() },
		"TimestampNano": func() int64 { return time.Now().UnixNano() },
	}
}

type templateData struct {
	VU   templateVU
	Vars map[string]any
	Env  func(string) string
}

type templateVU struct {
	ID        int
	Iteration int
}

func envMap(key string) string {
	return os.Getenv(key)
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func randomString(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	for i := range b {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return string(b)
}
