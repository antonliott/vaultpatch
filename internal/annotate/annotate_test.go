package annotate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/annotate"
)

type mockWriter struct {
	data    map[string]map[string]string
	readErr error
	writeErr error
	written map[string]map[string]string
}

func (m *mockWriter) Read(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
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

func (m *mockWriter) Write(_ context.Context, path string, data map[string]string) error {
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
	mw := &mockWriter{data: map[string]map[string]string{"secret/app": {"key": "val"}}}
	a := annotate.NewAnnotator(mw)
	res := a.Apply(context.Background(), annotate.Options{
		Path:        "secret/app",
		Annotations: map[string]string{"env": "prod"},
		DryRun:      true,
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Added != 1 {
		t.Errorf("expected 1 added, got %d", res.Added)
	}
	if mw.written != nil {
		t.Error("expected no writes on dry-run")
	}
}

func TestApply_AddsAndUpdates(t *testing.T) {
	mw := &mockWriter{data: map[string]map[string]string{"secret/app": {"env": "dev", "owner": "alice"}}}
	a := annotate.NewAnnotator(mw)
	res := a.Apply(context.Background(), annotate.Options{
		Path:        "secret/app",
		Annotations: map[string]string{"env": "prod", "version": "2"},
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Updated != 1 {
		t.Errorf("expected 1 updated, got %d", res.Updated)
	}
	if res.Added != 1 {
		t.Errorf("expected 1 added, got %d", res.Added)
	}
	if mw.written["secret/app"]["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", mw.written["secret/app"]["env"])
	}
}

func TestApply_RemovesKeys(t *testing.T) {
	mw := &mockWriter{data: map[string]map[string]string{"secret/app": {"env": "dev", "deprecated": "yes"}}}
	a := annotate.NewAnnotator(mw)
	res := a.Apply(context.Background(), annotate.Options{
		Path:       "secret/app",
		RemoveKeys: []string{"deprecated"},
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", res.Removed)
	}
	if _, ok := mw.written["secret/app"]["deprecated"]; ok {
		t.Error("expected deprecated key to be removed")
	}
}

func TestApply_ReadError(t *testing.T) {
	mw := &mockWriter{readErr: errors.New("vault unavailable")}
	a := annotate.NewAnnotator(mw)
	res := a.Apply(context.Background(), annotate.Options{Path: "secret/app"})
	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_WriteError(t *testing.T) {
	mw := &mockWriter{
		data:     map[string]map[string]string{"secret/app": {}},
		writeErr: errors.New("write failed"),
	}
	a := annotate.NewAnnotator(mw)
	res := a.Apply(context.Background(), annotate.Options{
		Path:        "secret/app",
		Annotations: map[string]string{"k": "v"},
	})
	if res.Err == nil {
		t.Fatal("expected write error, got nil")
	}
}
