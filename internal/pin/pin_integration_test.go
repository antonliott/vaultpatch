package pin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/pin"
)

// storeRW is an in-memory Reader/Writer for integration tests.
type storeRW struct {
	store map[string]map[string]string
}

func newStoreRW(initial map[string]map[string]string) *storeRW {
	if initial == nil {
		initial = make(map[string]map[string]string)
	}
	return &storeRW{store: initial}
}

func (s *storeRW) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	v, ok := s.store[path]
	if !ok {
		return map[string]string{}, nil
	}
	copy := make(map[string]string, len(v))
	for k, val := range v {
		copy[k] = val
	}
	return copy, nil
}

func (s *storeRW) WriteSecret(_ context.Context, path string, data map[string]string) error {
	s.store[path] = data
	return nil
}

func TestPinAndRestore_RoundTrip(t *testing.T) {
	store := newStoreRW(map[string]map[string]string{
		"secret/db": {"password": "hunter2", "host": "localhost"},
	})
	p := pin.NewPinner(store, store)
	ctx := context.Background()

	entry, err := p.Pin(ctx, "secret/db")
	if err != nil {
		t.Fatalf("pin: %v", err)
	}

	// Mutate live secret
	store.store["secret/db"]["password"] = "changed"
	delete(store.store["secret/db"], "host")

	result, err := p.CheckDrift(ctx, entry)
	if err != nil {
		t.Fatalf("drift: %v", err)
	}
	if !result.Drifted {
		t.Fatal("expected drift after mutation")
	}

	if err := p.Restore(ctx, entry, false); err != nil {
		t.Fatalf("restore: %v", err)
	}

	post, _ := p.CheckDrift(ctx, entry)
	if post.Drifted {
		t.Errorf("expected no drift after restore, got: %v", post.Diffs)
	}
}

func TestFormatDrift_ShowsAllStatuses(t *testing.T) {
	results := []*pin.DriftResult{
		{Path: "secret/a", Drifted: false},
		{Path: "secret/b", Drifted: true, Diffs: []string{"~ key (changed)"}},
	}
	out := pin.FormatDrift(results)
	if !strings.Contains(out, "ok") {
		t.Error("expected 'ok' for non-drifted path")
	}
	if !strings.Contains(out, "DRIFT") {
		t.Error("expected 'DRIFT' for drifted path")
	}
	if !strings.Contains(out, "~ key (changed)") {
		t.Error("expected diff line in output")
	}
}
