package compare_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/compare"
)

type mockReader struct {
	list func(ctx context.Context, path string) ([]string, error)
	read func(ctx context.Context, path string) (map[string]interface{}, error)
}

func (m *mockReader) ListSecrets(ctx context.Context, path string) ([]string, error) {
	return m.list(ctx, path)
}
func (m *mockReader) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	return m.read(ctx, path)
}

func staticReader(data map[string]map[string]interface{}) *mockReader {
	return &mockReader{
		list: func(_ context.Context, path string) ([]string, error) {
			var keys []string
			for k := range data {
				if len(k) > len(path) && k[:len(path)] == path {
					keys = append(keys, k[len(path)+1:])
				}
			}
			return keys, nil
		},
		read: func(_ context.Context, path string) (map[string]interface{}, error) {
			v, ok := data[path]
			if !ok {
				return nil, nil
			}
			return v, nil
		},
	}
}

func TestCompare_NoDiff(t *testing.T) {
	src := staticReader(map[string]map[string]interface{}{"kv/db": {"pass": "s3cr3t"}})
	tgt := staticReader(map[string]map[string]interface{}{"kv/db": {"pass": "s3cr3t"}})
	c := compare.NewComparator(src, tgt)
	r, err := c.Compare(context.Background(), "kv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.HasDiff() {
		t.Fatalf("expected no diff, got %+v", r.Deltas)
	}
}

func TestCompare_Added(t *testing.T) {
	src := staticReader(map[string]map[string]interface{}{"kv/db": {"pass": "new"}})
	tgt := staticReader(map[string]map[string]interface{}{"kv/db": {}})
	c := compare.NewComparator(src, tgt)
	r, err := c.Compare(context.Background(), "kv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Deltas) != 1 || r.Deltas[0].Kind != compare.Added {
		t.Fatalf("expected one Added delta, got %+v", r.Deltas)
	}
}

func TestCompare_ListError(t *testing.T) {
	src := &mockReader{
		list: func(_ context.Context, _ string) ([]string, error) {
			return nil, errors.New("list failed")
		},
		read: func(_ context.Context, _ string) (map[string]interface{}, error) { return nil, nil },
	}
	tgt := staticReader(nil)
	c := compare.NewComparator(src, tgt)
	_, err := c.Compare(context.Background(), "kv")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormat_NoDiff(t *testing.T) {
	out := compare.Format(compare.Result{})
	if out != "(no differences)" {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestFormat_WithDeltas(t *testing.T) {
	r := compare.Result{Deltas: []compare.Delta{
		{Path: "kv/db", Key: "pass", SourceVal: "new", Kind: compare.Added},
		{Path: "kv/db", Key: "old", TargetVal: "gone", Kind: compare.Removed},
	}}
	out := compare.Format(r)
	if out == "" || out == "(no differences)" {
		t.Fatalf("expected formatted output, got %q", out)
	}
}
