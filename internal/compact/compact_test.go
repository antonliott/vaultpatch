package compact_test

import (
	"context"
	"errors"
	"testing"

	"github.com/youorg/vaultpatch/internal/compact"
)

type mockRW struct {
	data     map[string]map[string]string
	written  map[string]map[string]string
	readErr  map[string]error
	writeErr map[string]error
}

func (m *mockRW) List(_ context.Context, _ string) ([]string, error) { return nil, nil }

func (m *mockRW) Read(_ context.Context, path string) (map[string]string, error) {
	if err := m.readErr[path]; err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for k, v := range m.data[path] {
		out[k] = v
	}
	return out, nil
}

func (m *mockRW) Write(_ context.Context, path string, data map[string]string) error {
	if err := m.writeErr[path]; err != nil {
		return err
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"a": {"x": "1", "y": "2"},
			"b": {"x": "9", "z": "3"},
		},
		written: map[string]map[string]string{},
	}
	c := compact.New(rw, true)
	results, err := c.Apply(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rw.written) != 0 {
		t.Error("expected no writes in dry-run")
	}
	if len(results[1].Removed) != 1 || results[1].Removed[0] != "x" {
		t.Errorf("expected x removed from b, got %v", results[1].Removed)
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"a": {"foo": "bar"},
			"b": {"foo": "baz", "new": "val"},
		},
		written: map[string]map[string]string{},
	}
	c := compact.New(rw, false)
	results, _ := c.Apply(context.Background(), []string{"a", "b"})
	if _, ok := rw.written["b"]; !ok {
		t.Error("expected b to be written")
	}
	if _, dup := rw.written["b"]["foo"]; dup {
		t.Error("expected foo removed from b")
	}
	_ = results
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{
		data:    map[string]map[string]string{},
		written: map[string]map[string]string{},
		readErr: map[string]error{"a": errors.New("vault down")},
	}
	c := compact.New(rw, false)
	results, _ := c.Apply(context.Background(), []string{"a"})
	if results[0].Err == nil {
		t.Error("expected error for path a")
	}
}
