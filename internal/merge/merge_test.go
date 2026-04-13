package merge_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultpatch/vaultpatch/internal/merge"
)

type mockRW struct {
	data    map[string]map[string]string
	written map[string]map[string]string
	readErr error
	writeErr error
}

func (m *mockRW) ReadSecrets(_ context.Context, path string) (map[string]string, error) {
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

func (m *mockRW) WriteSecrets(_ context.Context, path string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.written == nil {
		m.written = make(map[string]map[string]string)
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"src/a": {"key1": "val1"},
	}}
	m := merge.NewMerger(rw, merge.Options{DryRun: true})
	res, err := m.Apply(context.Background(), "dst", []string{"src/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Merged != 1 {
		t.Errorf("expected 1 merged, got %d", res.Merged)
	}
	if rw.written != nil {
		t.Error("dry run should not write")
	}
}

func TestApply_Success_NoOverwrite(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"dst":   {"shared": "original"},
		"src/a": {"shared": "new", "extra": "value"},
	}}
	m := merge.NewMerger(rw, merge.Options{})
	res, err := m.Apply(context.Background(), "dst", []string{"src/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", res.Skipped)
	}
	if rw.written["dst"]["shared"] != "original" {
		t.Errorf("expected original value preserved, got %q", rw.written["dst"]["shared"])
	}
}

func TestApply_Success_Overwrite(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"dst":   {"shared": "original"},
		"src/a": {"shared": "new"},
	}}
	m := merge.NewMerger(rw, merge.Options{Overwrite: true})
	res, err := m.Apply(context.Background(), "dst", []string{"src/a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Merged != 1 {
		t.Errorf("expected 1 merged, got %d", res.Merged)
	}
	if rw.written["dst"]["shared"] != "new" {
		t.Errorf("expected overwritten value, got %q", rw.written["dst"]["shared"])
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault unavailable")}
	m := merge.NewMerger(rw, merge.Options{})
	_, err := m.Apply(context.Background(), "dst", []string{"src/a"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_WriteError(t *testing.T) {
	rw := &mockRW{
		data:     map[string]map[string]string{"src/a": {"k": "v"}},
		writeErr: errors.New("write failed"),
	}
	m := merge.NewMerger(rw, merge.Options{})
	_, err := m.Apply(context.Background(), "dst", []string{"src/a"})
	if err == nil {
		t.Fatal("expected error on write failure")
	}
}
