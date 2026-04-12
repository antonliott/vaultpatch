package rollback

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/snapshot"
)

type mockWriter struct {
	store   map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (m *mockWriter) Read(_ context.Context, path string) (map[string]interface{}, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.store[path], nil
}

func (m *mockWriter) Write(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.store[path] = data
	return nil
}

func makeSnap(secrets map[string]map[string]interface{}) *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Namespace: "ns1",
		CreatedAt: time.Now(),
		Secrets:   secrets,
	}
}

func TestApply_DryRun(t *testing.T) {
	mw := &mockWriter{store: map[string]map[string]interface{}{
		"secret/a": {"k": "old"},
	}}
	r := NewRollbacker(mw, true)
	snap := makeSnap(map[string]map[string]interface{}{"secret/a": {"k": "new"}})
	res, err := r.Apply(context.Background(), snap)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.WouldRevert) != 1 {
		t.Errorf("expected 1 would-revert, got %d", len(res.WouldRevert))
	}
	if mw.store["secret/a"]["k"] != "old" {
		t.Error("dry-run must not modify store")
	}
}

func TestApply_Success(t *testing.T) {
	mw := &mockWriter{store: map[string]map[string]interface{}{
		"secret/a": {"k": "old"},
	}}
	r := NewRollbacker(mw, false)
	snap := makeSnap(map[string]map[string]interface{}{"secret/a": {"k": "snap"}})
	res, err := r.Apply(context.Background(), snap)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Reverted) != 1 {
		t.Errorf("expected 1 reverted, got %d", len(res.Reverted))
	}
}

func TestApply_SkipsUnchanged(t *testing.T) {
	mw := &mockWriter{store: map[string]map[string]interface{}{
		"secret/a": {"k": "same"},
	}}
	r := NewRollbacker(mw, false)
	snap := makeSnap(map[string]map[string]interface{}{"secret/a": {"k": "same"}})
	res, _ := r.Apply(context.Background(), snap)
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", res.Skipped)
	}
}

func TestApply_ReadError(t *testing.T) {
	mw := &mockWriter{store: map[string]map[string]interface{}{}, readErr: errors.New("vault down")}
	r := NewRollbacker(mw, false)
	snap := makeSnap(map[string]map[string]interface{}{"secret/a": {"k": "v"}})
	res, _ := r.Apply(context.Background(), snap)
	if !res.HasErrors() {
		t.Error("expected errors")
	}
}
