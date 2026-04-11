package rotate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/rotate"
)

type mockWriter struct {
	data    map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (m *mockWriter) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if v, ok := m.data[path]; ok {
		return v, nil
	}
	return map[string]interface{}{}, nil
}

func (m *mockWriter) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.data[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	w := &mockWriter{data: map[string]map[string]interface{}{}}
	r := rotate.NewRotator(w, true)
	reqs := []rotate.RotateRequest{{Path: "secret/foo", Updates: map[string]string{"key": "newval"}}}
	results := r.Apply(context.Background(), reqs)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("expected 1 result with no error, got %+v", results)
	}
	if _, written := w.data["secret/foo"]; written {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	w := &mockWriter{data: map[string]map[string]interface{}{"secret/foo": {"old": "val"}}}
	r := rotate.NewRotator(w, false)
	reqs := []rotate.RotateRequest{{Path: "secret/foo", Updates: map[string]string{"new": "rotated"}}}
	results := r.Apply(context.Background(), reqs)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if w.data["secret/foo"]["new"] != "rotated" {
		t.Error("expected rotated key to be written")
	}
	if w.data["secret/foo"]["old"] != "val" {
		t.Error("expected existing key to be preserved")
	}
}

func TestApply_ReadError(t *testing.T) {
	w := &mockWriter{data: map[string]map[string]interface{}{}, readErr: errors.New("vault down")}
	r := rotate.NewRotator(w, false)
	reqs := []rotate.RotateRequest{{Path: "secret/bar", Updates: map[string]string{"k": "v"}}}
	results := r.Apply(context.Background(), reqs)
	if results[0].Err == nil {
		t.Fatal("expected read error")
	}
}

func TestApply_WriteError(t *testing.T) {
	w := &mockWriter{data: map[string]map[string]interface{}{}, writeErr: errors.New("forbidden")}
	r := rotate.NewRotator(w, false)
	reqs := []rotate.RotateRequest{{Path: "secret/baz", Updates: map[string]string{"k": "v"}}}
	results := r.Apply(context.Background(), reqs)
	if results[0].Err == nil {
		t.Fatal("expected write error")
	}
}
