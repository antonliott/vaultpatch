package archive_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/archive"
)

type mockReader struct {
	data map[string]map[string]interface{}
	listErr error
	readErr error
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]interface{}, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.data[path], nil
}

func (m *mockReader) List(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

type mockWriter struct {
	written map[string]map[string]interface{}
	writeErr error
}

func (m *mockWriter) Write(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.written == nil {
		m.written = make(map[string]map[string]interface{})
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	r := &mockReader{data: map[string]map[string]interface{}{"app/db": {"pass": "s3cr3t"}}}
	w := &mockWriter{}
	a := archive.NewArchiver(r, w, "archive", true)
	res, err := a.Apply(context.Background(), []string{"app/db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Archived) != 1 {
		t.Errorf("expected 1 archived, got %d", len(res.Archived))
	}
	if len(w.written) != 0 {
		t.Error("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	r := &mockReader{data: map[string]map[string]interface{}{"app/db": {"pass": "s3cr3t"}}}
	w := &mockWriter{}
	a := archive.NewArchiver(r, w, "archive", false)
	res, err := a.Apply(context.Background(), []string{"app/db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Archived) != 1 || res.Archived[0] != "app/db" {
		t.Errorf("unexpected archived list: %v", res.Archived)
	}
	if _, ok := w.written["archive/app/db"]; !ok {
		t.Error("expected secret written to archive path")
	}
	if _, ok := w.written["archive/app/db"]["_archived_at"]; !ok {
		t.Error("expected _archived_at metadata key")
	}
}

func TestApply_SkipsNilSecret(t *testing.T) {
	r := &mockReader{data: map[string]map[string]interface{}{}}
	w := &mockWriter{}
	a := archive.NewArchiver(r, w, "archive", false)
	res, err := a.Apply(context.Background(), []string{"app/missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(res.Skipped))
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{readErr: errors.New("vault unavailable")}
	w := &mockWriter{}
	a := archive.NewArchiver(r, w, "archive", false)
	res, err := a.Apply(context.Background(), []string{"app/db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasErrors() {
		t.Error("expected errors in result")
	}
}

func TestApply_WriteError(t *testing.T) {
	r := &mockReader{data: map[string]map[string]interface{}{"app/db": {"pass": "x"}}}
	w := &mockWriter{writeErr: errors.New("write failed")}
	a := archive.NewArchiver(r, w, "archive", false)
	res, err := a.Apply(context.Background(), []string{"app/db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasErrors() {
		t.Error("expected write error in result")
	}
}
