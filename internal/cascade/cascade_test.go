package cascade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/cascade"
)

type mockRW struct {
	data    map[string]map[string]string
	written map[string]map[string]string
	readErr map[string]error
}

func newMockRW() *mockRW {
	return &mockRW{
		data:    make(map[string]map[string]string),
		written: make(map[string]map[string]string),
		readErr: make(map[string]error),
	}
}

func (m *mockRW) Read(_ context.Context, path string) (map[string]string, error) {
	if err, ok := m.readErr[path]; ok {
		return nil, err
	}
	if d, ok := m.data[path]; ok {
		copy := make(map[string]string, len(d))
		for k, v := range d {
			copy[k] = v
		}
		return copy, nil
	}
	return map[string]string{}, nil
}

func (m *mockRW) Write(_ context.Context, path string, data map[string]string) error {
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"key": "val"}
	c := cascade.NewCascader(rw, cascade.Options{DryRun: true})
	results := c.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if _, ok := rw.written["secret/child"]; ok {
		t.Fatal("dry-run should not write")
	}
}

func TestApply_Success_NoOverwrite(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"a": "1", "b": "2"}
	rw.data["secret/child"] = map[string]string{"a": "existing"}
	c := cascade.NewCascader(rw, cascade.Options{})
	results := c.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if results[0].Written != 1 {
		t.Errorf("expected 1 written, got %d", results[0].Written)
	}
	if results[0].Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", results[0].Skipped)
	}
	if rw.written["secret/child"]["a"] != "existing" {
		t.Error("existing key should not be overwritten")
	}
}

func TestApply_Success_Overwrite(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"a": "new"}
	rw.data["secret/child"] = map[string]string{"a": "old"}
	c := cascade.NewCascader(rw, cascade.Options{Overwrite: true})
	results := c.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if rw.written["secret/child"]["a"] != "new" {
		t.Error("key should be overwritten")
	}
}

func TestApply_ParentReadError(t *testing.T) {
	rw := newMockRW()
	rw.readErr["secret/parent"] = errors.New("permission denied")
	c := cascade.NewCascader(rw, cascade.Options{})
	results := c.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if results[0].Err == nil {
		t.Fatal("expected error from parent read failure")
	}
}

func TestApply_NoChanges_SkipsWrite(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"x": "1"}
	rw.data["secret/child"] = map[string]string{"x": "1"}
	c := cascade.NewCascader(rw, cascade.Options{})
	results := c.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if results[0].Written != 0 {
		t.Errorf("expected 0 written, got %d", results[0].Written)
	}
	if _, ok := rw.written["secret/child"]; ok {
		t.Fatal("should not write when nothing changed")
	}
}
