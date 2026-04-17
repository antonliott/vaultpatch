package shield_test

import (
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/shield"
)

type mockRW struct {
	data    map[string]map[string]string
	readErr error
	writeErr error
	written map[string]map[string]string
}

func (m *mockRW) Read(path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if v, ok := m.data[path]; ok {
		copy := make(map[string]string, len(v))
		for k, val := range v {
			copy[k] = val
		}
		return copy, nil
	}
	return map[string]string{}, nil
}

func (m *mockRW) Write(path string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.written == nil {
		m.written = map[string]map[string]string{}
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"sec/a": {"key": "old"}}}
	s := shield.NewShielder(rw, []string{"key"}, true)
	res, err := s.Apply("sec/a", map[string]string{"key": "new", "other": "val"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Blocked {
		t.Error("expected blocked")
	}
	if len(res.Protected) != 1 || res.Protected[0] != "key" {
		t.Errorf("unexpected protected: %v", res.Protected)
	}
	if rw.written != nil {
		t.Error("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"sec/a": {"locked": "keep", "free": "old"}}}
	s := shield.NewShielder(rw, []string{"locked"}, false)
	res, err := s.Apply("sec/a", map[string]string{"locked": "changed", "free": "new"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Blocked {
		t.Error("expected blocked")
	}
	written := rw.written["sec/a"]
	if written["locked"] != "keep" {
		t.Errorf("protected key mutated: %s", written["locked"])
	}
	if written["free"] != "new" {
		t.Errorf("free key not updated: %s", written["free"])
	}
}

func TestApply_NoProtectedKeys(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"sec/b": {"a": "1"}}}
	s := shield.NewShielder(rw, nil, false)
	res, err := s.Apply("sec/b", map[string]string{"a": "2", "b": "3"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Blocked {
		t.Error("expected no block")
	}
	if rw.written["sec/b"]["a"] != "2" {
		t.Error("expected a=2")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault down")}
	s := shield.NewShielder(rw, []string{"x"}, false)
	_, err := s.Apply("sec/c", map[string]string{"x": "v"})
	if err == nil {
		t.Error("expected error")
	}
}
