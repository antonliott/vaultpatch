package rename_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/rename"
)

type mockRW struct {
	data    map[string]string
	readErr error
	writeErr error
	written map[string]string
}

func (m *mockRW) ReadSecret(_ context.Context, _ string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	out := make(map[string]string, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockRW) WriteSecret(_ context.Context, _ string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]string{"old_key": "value"}}
	r := rename.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", "old_key", "new_key", true)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if rw.written != nil {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{data: map[string]string{"old_key": "value", "other": "x"}}
	r := rename.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", "old_key", "new_key", false)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := rw.written["new_key"]; !ok {
		t.Error("expected new_key in written data")
	}
	if _, ok := rw.written["old_key"]; ok {
		t.Error("expected old_key to be removed")
	}
}

func TestApply_KeyMissing_Skipped(t *testing.T) {
	rw := &mockRW{data: map[string]string{"other": "x"}}
	r := rename.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", "old_key", "new_key", false)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Skipped {
		t.Error("expected result to be skipped")
	}
}

func TestApply_NewKeyExists_Error(t *testing.T) {
	rw := &mockRW{data: map[string]string{"old_key": "a", "new_key": "b"}}
	r := rename.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", "old_key", "new_key", false)
	if res.Err == nil {
		t.Fatal("expected error when new key already exists")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault unavailable")}
	r := rename.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", "old_key", "new_key", false)
	if res.Err == nil {
		t.Fatal("expected error on read failure")
	}
}
