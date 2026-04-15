package prefix_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/prefix"
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
	rw := &mockRW{data: map[string]string{"OLD_FOO": "bar", "OTHER": "val"}}
	r := prefix.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", prefix.Options{
		OldPrefix: "OLD_",
		NewPrefix: "NEW_",
		DryRun:    true,
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if _, ok := res.Renamed["OLD_FOO"]; !ok {
		t.Error("expected OLD_FOO in Renamed")
	}
	if rw.written != nil {
		t.Error("expected no write in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{data: map[string]string{"DEV_HOST": "localhost", "DEV_PORT": "5432", "SHARED": "x"}}
	r := prefix.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/db", prefix.Options{
		OldPrefix: "DEV_",
		NewPrefix: "PROD_",
	})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Renamed) != 2 {
		t.Errorf("expected 2 renamed keys, got %d", len(res.Renamed))
	}
	if rw.written["PROD_HOST"] != "localhost" {
		t.Errorf("expected PROD_HOST=localhost, got %q", rw.written["PROD_HOST"])
	}
	if rw.written["SHARED"] != "x" {
		t.Error("expected SHARED key preserved")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("permission denied")}
	r := prefix.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", prefix.Options{OldPrefix: "A_", NewPrefix: "B_"})
	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_NoMatchingKeys(t *testing.T) {
	rw := &mockRW{data: map[string]string{"FOO": "1", "BAR": "2"}}
	r := prefix.NewRenamer(rw)
	res := r.Apply(context.Background(), "secret/app", prefix.Options{OldPrefix: "MISSING_", NewPrefix: "X_"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Renamed) != 0 {
		t.Errorf("expected 0 renames, got %d", len(res.Renamed))
	}
	if rw.written != nil {
		t.Error("expected no write when no keys match")
	}
}
