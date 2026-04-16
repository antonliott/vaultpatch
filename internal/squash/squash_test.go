package squash_test

import (
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/squash"
)

type mockRW struct {
	store   map[string]map[string]string
	written map[string]map[string]string
	readErr error
}

func (m *mockRW) Read(path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if d, ok := m.store[path]; ok {
		out := make(map[string]string, len(d))
		for k, v := range d {
			out[k] = v
		}
		return out, nil
	}
	return map[string]string{}, nil
}

func (m *mockRW) Write(path string, data map[string]string) error {
	if m.written == nil {
		m.written = make(map[string]map[string]string)
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{
		"a": {"x": "1"},
		"b": {"y": "2"},
	}}
	s := squash.New(rw, true)
	r := s.Apply("dest", []string{"a", "b"})
	if !r.DryRun {
		t.Fatal("expected dry-run")
	}
	if r.Merged != 2 {
		t.Fatalf("expected 2 merged, got %d", r.Merged)
	}
	if rw.written != nil {
		t.Fatal("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{
		"a": {"x": "1", "shared": "from-a"},
		"b": {"y": "2", "shared": "from-b"},
	}}
	s := squash.New(rw, false)
	r := s.Apply("dest", []string{"a", "b"})
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Merged != 2 {
		t.Fatalf("expected 2, got %d", r.Merged)
	}
	got := rw.written["dest"]
	if got["shared"] != "from-b" {
		t.Errorf("expected later source to win, got %s", got["shared"])
	}
	if got["x"] != "1" || got["y"] != "2" {
		t.Error("missing keys from sources")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault down")}
	s := squash.New(rw, false)
	r := s.Apply("dest", []string{"a"})
	if r.Err == nil {
		t.Fatal("expected error")
	}
}

func TestResult_String_DryRun(t *testing.T) {
	r := squash.Result{Destination: "out", Merged: 3, DryRun: true}
	got := r.String()
	if got == "" {
		t.Fatal("empty string")
	}
}
