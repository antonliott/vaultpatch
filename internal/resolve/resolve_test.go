package resolve_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/resolve"
)

type mockReader struct {
	data map[string]map[string]string
	err  error
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	if d, ok := m.data[path]; ok {
		return d, nil
	}
	return map[string]string{}, nil
}

func TestApply_NoReferences(t *testing.T) {
	r := resolve.New(&mockReader{data: map[string]map[string]string{
		"secret/app": {"host": "localhost", "port": "5432"},
	}})
	out, n, err := r.Apply(context.Background(), "secret/app", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 resolutions, got %d", n)
	}
	if out["host"] != "localhost" {
		t.Errorf("expected localhost, got %q", out["host"])
	}
}

func TestApply_ResolvesReference(t *testing.T) {
	r := resolve.New(&mockReader{data: map[string]map[string]string{
		"secret/app": {"dsn": "postgres://{{secret/db#user}}:{{secret/db#pass}}@localhost/app"},
		"secret/db":  {"user": "admin", "pass": "s3cr3t"},
	}})
	out, n, err := r.Apply(context.Background(), "secret/app", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 resolutions, got %d", n)
	}
	want := "postgres://admin:s3cr3t@localhost/app"
	if out["dsn"] != want {
		t.Errorf("expected %q, got %q", want, out["dsn"])
	}
}

func TestApply_MissingRefKey(t *testing.T) {
	r := resolve.New(&mockReader{data: map[string]map[string]string{
		"secret/app": {"val": "{{secret/db#missing}}"},
		"secret/db":  {"user": "admin"},
	}})
	_, _, err := r.Apply(context.Background(), "secret/app", false)
	if err == nil {
		t.Fatal("expected error for missing ref key")
	}
}

func TestApply_ReadError(t *testing.T) {
	r := resolve.New(&mockReader{err: errors.New("vault unavailable")})
	_, _, err := r.Apply(context.Background(), "secret/app", false)
	if err == nil {
		t.Fatal("expected error on read failure")
	}
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	orig := map[string]string{"key": "{{secret/other#val}}"}
	store := map[string]map[string]string{
		"secret/app":   orig,
		"secret/other": {"val": "replaced"},
	}
	r := resolve.New(&mockReader{data: store})
	out, _, err := r.Apply(context.Background(), "secret/app", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["key"] == orig["key"] {
		t.Error("expected output to differ from original placeholder")
	}
	if orig["key"] != "{{secret/other#val}}" {
		t.Error("original map was mutated")
	}
}
