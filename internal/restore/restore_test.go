package restore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/restore"
	"github.com/example/vaultpatch/internal/snapshot"
)

type mockWriter struct {
	failPaths map[string]bool
	written   []string
}

func (m *mockWriter) Write(_ context.Context, path string, _ map[string]interface{}) error {
	if m.failPaths[path] {
		return errors.New("vault write error")
	}
	m.written = append(m.written, path)
	return nil
}

func makeSnap() *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Namespace: "ns1",
		TakenAt:   time.Now(),
		Secrets: map[string]map[string]interface{}{
			"secret/foo": {"key": "val"},
			"secret/bar": {"pass": "s3cr3t"},
		},
	}
}

func TestApply_DryRun(t *testing.T) {
	w := &mockWriter{}
	r := restore.NewRestorer(w, true)
	result, err := r.Apply(context.Background(), makeSnap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Restored) != 0 {
		t.Errorf("expected 0 restored, got %d", len(result.Restored))
	}
	if len(result.Skipped) != 2 {
		t.Errorf("expected 2 skipped, got %d", len(result.Skipped))
	}
	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestApply_Success(t *testing.T) {
	w := &mockWriter{}
	r := restore.NewRestorer(w, false)
	result, err := r.Apply(context.Background(), makeSnap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Restored) != 2 {
		t.Errorf("expected 2 restored, got %d", len(result.Restored))
	}
	if result.HasErrors() {
		t.Error("expected no errors")
	}
}

func TestApply_PartialError(t *testing.T) {
	w := &mockWriter{failPaths: map[string]bool{"secret/foo": true}}
	r := restore.NewRestorer(w, false)
	result, _ := r.Apply(context.Background(), makeSnap())
	if !result.HasErrors() {
		t.Error("expected errors")
	}
	if len(result.Restored) != 1 {
		t.Errorf("expected 1 restored, got %d", len(result.Restored))
	}
}

func TestResult_Summary(t *testing.T) {
	r := &restore.RestoreResult{Restored: []string{"a", "b"}, DryRun: false}
	s := r.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}
