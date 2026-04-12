package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/sync"
)

// mockReader implements SecretReader.
type mockReader struct {
	keys    []string
	listErr error
	readErr error
	data    map[string]interface{}
}

func (m *mockReader) ListSecrets(_ context.Context, _, _ string) ([]string, error) {
	return m.keys, m.listErr
}

func (m *mockReader) ReadSecret(_ context.Context, _, _ string) (map[string]interface{}, error) {
	return m.data, m.readErr
}

// mockWriter implements SecretWriter.
type mockWriter struct {
	written  []string
	writeErr error
}

func (m *mockWriter) WriteSecret(_ context.Context, _, path string, _ map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = append(m.written, path)
	return nil
}

func TestApply_DryRun(t *testing.T) {
	reader := &mockReader{keys: []string{"secret/a", "secret/b"}, data: map[string]interface{}{"k": "v"}}
	writer := &mockWriter{}
	s := sync.NewSyncer(reader, writer)

	res := s.Apply(context.Background(), sync.Options{Mount: "secret", DryRun: true})

	if len(res.Synced) != 2 {
		t.Fatalf("expected 2 synced, got %d", len(res.Synced))
	}
	if len(writer.written) != 0 {
		t.Fatal("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	reader := &mockReader{keys: []string{"secret/a"}, data: map[string]interface{}{"k": "v"}}
	writer := &mockWriter{}
	s := sync.NewSyncer(reader, writer)

	res := s.Apply(context.Background(), sync.Options{Mount: "secret"})

	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(writer.written) != 1 {
		t.Fatalf("expected 1 write, got %d", len(writer.written))
	}
}

func TestApply_ListError(t *testing.T) {
	reader := &mockReader{listErr: errors.New("permission denied")}
	writer := &mockWriter{}
	s := sync.NewSyncer(reader, writer)

	res := s.Apply(context.Background(), sync.Options{Mount: "secret"})

	if !res.HasErrors() {
		t.Fatal("expected error")
	}
}

func TestApply_ReadError(t *testing.T) {
	reader := &mockReader{keys: []string{"secret/a"}, readErr: errors.New("read fail")}
	writer := &mockWriter{}
	s := sync.NewSyncer(reader, writer)

	res := s.Apply(context.Background(), sync.Options{Mount: "secret"})

	if !res.HasErrors() {
		t.Fatal("expected error from read")
	}
	if len(res.Synced) != 0 {
		t.Fatal("nothing should be synced on read error")
	}
}

func TestApply_WriteError(t *testing.T) {
	reader := &mockReader{keys: []string{"secret/a"}, data: map[string]interface{}{"k": "v"}}
	writer := &mockWriter{writeErr: errors.New("write fail")}
	s := sync.NewSyncer(reader, writer)

	res := s.Apply(context.Background(), sync.Options{Mount: "secret"})

	if !res.HasErrors() {
		t.Fatal("expected write error")
	}
}
