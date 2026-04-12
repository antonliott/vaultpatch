// Package template provides secret templating: render Go text/template strings
// against a map of Vault secret values.
package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// Renderer renders templates against a secrets map.
type Renderer struct {
	funcs template.FuncMap
}

// NewRenderer returns a Renderer with a default FuncMap.
func NewRenderer() *Renderer {
	return &Renderer{
		funcs: template.FuncMap{
			"required": requiredFn,
			"default":  defaultFn,
		},
	}
}

// Render executes tmplStr with the provided data map and returns the result.
func (r *Renderer) Render(name, tmplStr string, data map[string]string) (string, error) {
	t, err := template.New(name).Funcs(r.funcs).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	return buf.String(), nil
}

// RenderAll renders multiple named templates with the same data map.
// It returns a map of name -> rendered output, or the first error encountered.
func (r *Renderer) RenderAll(templates map[string]string, data map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(templates))
	for name, tmpl := range templates {
		result, err := r.Render(name, tmpl, data)
		if err != nil {
			return nil, fmt.Errorf("rendering %q: %w", name, err)
		}
		out[name] = result
	}
	return out, nil
}

func requiredFn(key string, data map[string]string) (string, error) {
	v, ok := data[key]
	if !ok || v == "" {
		return "", fmt.Errorf("required key %q is missing or empty", key)
	}
	return v, nil
}

func defaultFn(key, fallback string, data map[string]string) string {
	if v, ok := data[key]; ok && v != "" {
		return v
	}
	return fallback
}
