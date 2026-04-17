package immute_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/immute"
)

type mockRW struct {
	store map[string]map[string]string
	fail  bool
}

func (m *mockRW) Read(_ context.Context, path string) (map[string]string, error) {
	if m.fail {
		return nil, errors.New("read error")
	}
	if d, ok := m.store[path]; ok {
		out := make(map[string]string, len(d))
		for k, v := range d {
			out[k] = v
		}
		return out, nil
	}
	return map[string]string{}, nil
}

func (m *mockRW) Write(_ context.Context, path string, data map[string]string) error {
	if m.fail {
		return errors.New("write error")
	}
	m.store[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{"sec/a": {"key": "val"}}}
	l := immute.NewLocker(rw, true)
	res, err := l.Apply(context.Background(), []string{"sec/a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Marked) != 1 || res.Marked[0] != "sec/a" {
		t.Errorf("expected marked sec/a, got %v", res.Marked)
	}
	if _, ok := rw.store["sec/a"][immute.ImmutableKey]; ok {
		t.Error("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{"sec/a": {"key": "val"}}}
	l := immute.NewLocker(rw, false)
	res, err := l.Apply(context.Background(), []string{"sec/a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Marked) != 1 {
		t.Fatalf("expected 1 marked, got %d", len(res.Marked))
	}
	if rw.store["sec/a"][immute.ImmutableKey] != "true" {
		t.Error("expected _immutable=true written")
	}
}

func TestApply_SkipsAlreadyImmutable(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{"sec/a": {immute.ImmutableKey: "true"}}}
	l := immute.NewLocker(rw, false)
	res, err := l.Apply(context.Background(), []string{"sec/a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %v", res.Skipped)
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{store: map[string]map[string]string{}, fail: true}
	l := immute.NewLocker(rw, false)
	res, err := l.Apply(context.Background(), []string{"sec/a"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasErrors() {
		t.Error("expected errors")
	}
}

func TestResult_Summary_DryRun(t *testing.T) {
	res := immute.Result{DryRun: true, Marked: []string{"a", "b"}, Skipped: []string{"c"}}
	s := res.Summary()
	if s != "[dry-run] marked=2 skipped=1 errors=0" {
		t.Errorf("unexpected summary: %s", s)
	}
}
