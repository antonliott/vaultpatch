package clone_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/clone"
)

type mockReader struct {
	data map[string]string
	err  error
}

func (m *mockReader) ReadSecret(_ context.Context, _ string) (map[string]string, error) {
	return m.data, m.err
}

type mockWriter struct {
	called bool
	got    map[string]string
	err    error
}

func (m *mockWriter) WriteSecret(_ context.Context, _ string, data map[string]string) error {
	m.called = true
	m.got = data
	return m.err
}

func TestApply_DryRun(t *testing.T) {
	r := &mockReader{data: map[string]string{"key": "val"}}
	w := &mockWriter{}
	c := clone.NewCloner(r, w, true)
	res := c.Apply(context.Background(), "src/path", "dst/path")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if w.called {
		t.Fatal("writer should not be called in dry-run mode")
	}
	if !res.DryRun {
		t.Fatal("expected DryRun=true")
	}
}

func TestApply_Success(t *testing.T) {
	r := &mockReader{data: map[string]string{"a": "1", "b": "2"}}
	w := &mockWriter{}
	c := clone.NewCloner(r, w, false)
	res := c.Apply(context.Background(), "src/path", "dst/path")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !w.called {
		t.Fatal("expected writer to be called")
	}
	if len(res.Keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(res.Keys))
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{err: errors.New("vault unavailable")}
	w := &mockWriter{}
	c := clone.NewCloner(r, w, false)
	res := c.Apply(context.Background(), "src/path", "dst/path")
	if res.Err == nil {
		t.Fatal("expected error")
	}
	if w.called {
		t.Fatal("writer should not be called on read error")
	}
}

func TestApply_WriteError(t *testing.T) {
	r := &mockReader{data: map[string]string{"x": "y"}}
	w := &mockWriter{err: errors.New("permission denied")}
	c := clone.NewCloner(r, w, false)
	res := c.Apply(context.Background(), "src/path", "dst/path")
	if res.Err == nil {
		t.Fatal("expected error")
	}
}

func TestSummary_DryRun(t *testing.T) {
	res := clone.Result{SourcePath: "a", DestPath: "b", Keys: []string{"k"}, DryRun: true}
	s := res.Summary()
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}
