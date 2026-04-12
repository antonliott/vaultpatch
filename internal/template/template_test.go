package template

import (
	"strings"
	"testing"
)

func TestRender_SimpleLookup(t *testing.T) {
	r := NewRenderer()
	got, err := r.Render("t", `hello {{index . "name"}}`, map[string]string{"name": "vault"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello vault" {
		t.Errorf("got %q, want %q", got, "hello vault")
	}
}

func TestRender_Required_Present(t *testing.T) {
	r := NewRenderer()
	got, err := r.Render("t", `{{required "db_pass" .}}`, map[string]string{"db_pass": "s3cr3t"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "s3cr3t" {
		t.Errorf("got %q", got)
	}
}

func TestRender_Required_Missing(t *testing.T) {
	r := NewRenderer()
	_, err := r.Render("t", `{{required "db_pass" .}}`, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing required key")
	}
	if !strings.Contains(err.Error(), "db_pass") {
		t.Errorf("error should mention key name, got: %v", err)
	}
}

func TestRender_Default_UsedWhenMissing(t *testing.T) {
	r := NewRenderer()
	got, err := r.Render("t", `{{default "port" "5432" .}}`, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "5432" {
		t.Errorf("got %q, want %q", got, "5432")
	}
}

func TestRender_Default_OverriddenByValue(t *testing.T) {
	r := NewRenderer()
	got, err := r.Render("t", `{{default "port" "5432" .}}`, map[string]string{"port": "3306"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "3306" {
		t.Errorf("got %q, want %q", got, "3306")
	}
}

func TestRender_ParseError(t *testing.T) {
	r := NewRenderer()
	_, err := r.Render("t", `{{unclosed`, map[string]string{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestRenderAll_Success(t *testing.T) {
	r := NewRenderer()
	templates := map[string]string{
		"a": `val={{index . "key"}}`,
		"b": `KEY={{index . "key"}}`,
	}
	out, err := r.RenderAll(templates, map[string]string{"key": "X"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["a"] != "val=X" {
		t.Errorf("a: got %q", out["a"])
	}
	if out["b"] != "KEY=X" {
		t.Errorf("b: got %q", out["b"])
	}
}

func TestRenderAll_Error(t *testing.T) {
	r := NewRenderer()
	_, err := r.RenderAll(map[string]string{"bad": `{{bad syntax`}, map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}
