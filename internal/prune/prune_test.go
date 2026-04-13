package prune_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/prune"
)

type mockRW struct {
	data    map[string]interface{}
	readErr error
	writeErr error
	written map[string]interface{}
}

func (m *mockRW) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return m.data, m.readErr
}

func (m *mockRW) WriteSecret(_ context.Context, _ string, data map[string]interface{}) error {
	m.written = data
	return m.writeErr
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]interface{}{"password": "s3cr3t", "user": "admin"}}
	p := prune.NewPruner(rw, []string{"password"}, true)
	res := p.Apply(context.Background(), "secret/app")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.PrunedKeys) != 1 || res.PrunedKeys[0] != "password" {
		t.Errorf("expected pruned key 'password', got %v", res.PrunedKeys)
	}
	if rw.written != nil {
		t.Error("expected no write in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{data: map[string]interface{}{"token": "abc", "host": "localhost"}}
	p := prune.NewPruner(rw, []string{"token"}, false)
	res := p.Apply(context.Background(), "secret/db")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.PrunedKeys) != 1 {
		t.Fatalf("expected 1 pruned key, got %d", len(res.PrunedKeys))
	}
	if _, ok := rw.written["token"]; ok {
		t.Error("pruned key 'token' should not appear in written data")
	}
	if _, ok := rw.written["host"]; !ok {
		t.Error("key 'host' should remain in written data")
	}
}

func TestApply_NoMatchingKeys(t *testing.T) {
	rw := &mockRW{data: map[string]interface{}{"user": "bob"}}
	p := prune.NewPruner(rw, []string{"password"}, false)
	res := p.Apply(context.Background(), "secret/app")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.PrunedKeys) != 0 {
		t.Errorf("expected no pruned keys, got %v", res.PrunedKeys)
	}
	if rw.written != nil {
		t.Error("expected no write when nothing pruned")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("permission denied")}
	p := prune.NewPruner(rw, []string{"key"}, false)
	res := p.Apply(context.Background(), "secret/locked")

	if res.Err == nil {
		t.Fatal("expected error on read failure")
	}
}

func TestApply_WriteError(t *testing.T) {
	rw := &mockRW{
		data:     map[string]interface{}{"secret": "val"},
		writeErr: errors.New("vault unavailable"),
	}
	p := prune.NewPruner(rw, []string{"secret"}, false)
	res := p.Apply(context.Background(), "secret/svc")

	if res.Err == nil {
		t.Fatal("expected error on write failure")
	}
}
