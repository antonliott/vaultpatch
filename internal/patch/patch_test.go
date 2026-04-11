package patch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/youorg/vaultpatch/internal/diff"
	"github.com/youorg/vaultpatch/internal/patch"
)

type mockWriter struct {
	written map[string]map[string]interface{}
	deleted []string
	writeErr error
	deleteErr error
}

func newMockWriter() *mockWriter {
	return &mockWriter{written: make(map[string]map[string]interface{})}
}

func (m *mockWriter) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written[path] = data
	return nil
}

func (m *mockWriter) DeleteSecret(_ context.Context, path string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleted = append(m.deleted, path)
	return nil
}

func TestApply_NoChanges(t *testing.T) {
	w := newMockWriter()
	a := patch.NewApplier(w, false)
	n, err := a.Apply(context.Background(), "secret/app", nil)
	if err != nil || n != 0 {
		t.Fatalf("expected 0 changes, got %d err=%v", n, err)
	}
}

func TestApply_DryRun(t *testing.T) {
	w := newMockWriter()
	a := patch.NewApplier(w, true)
	changes := []diff.Change{{Type: diff.Added, Key: "foo", NewValue: "bar"}}
	n, err := a.Apply(context.Background(), "secret/app", changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	if len(w.written) != 0 {
		t.Fatal("dry-run should not write")
	}
}

func TestApply_WriteError(t *testing.T) {
	w := newMockWriter()
	w.writeErr = errors.New("vault unavailable")
	a := patch.NewApplier(w, false)
	changes := []diff.Change{{Type: diff.Added, Key: "foo", NewValue: "bar"}}
	_, err := a.Apply(context.Background(), "secret/app", changes)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApply_AddedAndRemoved(t *testing.T) {
	w := newMockWriter()
	a := patch.NewApplier(w, false)
	changes := []diff.Change{
		{Type: diff.Added, Key: "newkey", NewValue: "newval"},
		{Type: diff.Removed, Key: "oldkey", OldValue: "oldval"},
	}
	n, err := a.Apply(context.Background(), "secret/app", changes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 applied, got %d", n)
	}
}
