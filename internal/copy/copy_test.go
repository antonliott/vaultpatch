package copy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/copy"
)

type mockReader struct {
	data map[string]string
	err  error
}

func (m *mockReader) Read(_ context.Context, _ string) (map[string]string, error) {
	return m.data, m.err
}

type mockWriter struct {
	called bool
	last   map[string]string
	err    error
}

func (m *mockWriter) Write(_ context.Context, _ string, data map[string]string) error {
	m.called = true
	m.last = data
	return m.err
}

func TestApply_DryRun(t *testing.T) {
	r := &mockReader{data: map[string]string{"k": "v"}}
	w := &mockWriter{}
	c := copy.NewCopier(r, w, true)

	res := c.Apply(context.Background(), "src/path", "dst/path")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if w.called {
		t.Fatal("writer should not be called in dry-run mode")
	}
	if res.KeysCopied != 1 {
		t.Fatalf("expected 1 key, got %d", res.KeysCopied)
	}
	if !res.DryRun {
		t.Fatal("expected DryRun to be true")
	}
}

func TestApply_Success(t *testing.T) {
	data := map[string]string{"a": "1", "b": "2"}
	r := &mockReader{data: data}
	w := &mockWriter{}
	c := copy.NewCopier(r, w, false)

	res := c.Apply(context.Background(), "src/path", "dst/path")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !w.called {
		t.Fatal("expected writer to be called")
	}
	if res.KeysCopied != 2 {
		t.Fatalf("expected 2 keys, got %d", res.KeysCopied)
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{err: errors.New("vault unavailable")}
	w := &mockWriter{}
	c := copy.NewCopier(r, w, false)

	res := c.Apply(context.Background(), "src/path", "dst/path")

	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if w.called {
		t.Fatal("writer should not be called on read error")
	}
}

func TestApply_WriteError(t *testing.T) {
	r := &mockReader{data: map[string]string{"x": "y"}}
	w := &mockWriter{err: errors.New("permission denied")}
	c := copy.NewCopier(r, w, false)

	res := c.Apply(context.Background(), "src/path", "dst/path")

	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResult_Summary_DryRun(t *testing.T) {
	r := copy.Result{Source: "a", Destination: "b", KeysCopied: 3, DryRun: true}
	s := r.Summary()
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestResult_Summary_Error(t *testing.T) {
	r := copy.Result{Source: "a", Destination: "b", Err: errors.New("oops")}
	s := r.Summary()
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}
