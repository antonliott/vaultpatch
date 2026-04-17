package census_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/census"
)

type mockReader struct {
	listFn func(ctx context.Context, path string) ([]string, error)
	readFn func(ctx context.Context, path string) (map[string]string, error)
}

func (m *mockReader) List(ctx context.Context, path string) ([]string, error) {
	return m.listFn(ctx, path)
}
func (m *mockReader) Read(ctx context.Context, path string) (map[string]string, error) {
	return m.readFn(ctx, path)
}

func TestCollect_Success(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"alpha", "beta"}, nil
		},
		readFn: func(_ context.Context, path string) (map[string]string, error) {
			if path == "secrets/alpha" {
				return map[string]string{"host": "a", "port": "5432"}, nil
			}
			return map[string]string{"host": "b"}, nil
		},
	}

	c := census.NewCollector(r)
	rep, err := c.Collect(context.Background(), "secrets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.TotalPaths != 2 {
		t.Errorf("want 2 paths, got %d", rep.TotalPaths)
	}
	if rep.TotalKeys != 3 {
		t.Errorf("want 3 total keys, got %d", rep.TotalKeys)
	}
	if rep.KeyFreq["host"] != 2 {
		t.Errorf("want host freq 2, got %d", rep.KeyFreq["host"])
	}
	if rep.KeyFreq["port"] != 1 {
		t.Errorf("want port freq 1, got %d", rep.KeyFreq["port"])
	}
}

func TestCollect_ListError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return nil, errors.New("permission denied")
		},
	}
	c := census.NewCollector(r)
	_, err := c.Collect(context.Background(), "secrets")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_ReadError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"alpha"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, errors.New("read failed")
		},
	}
	c := census.NewCollector(r)
	_, err := c.Collect(context.Background(), "secrets")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_Empty(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{}, nil
		},
	}
	c := census.NewCollector(r)
	rep, err := c.Collect(context.Background(), "secrets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.TotalPaths != 0 || rep.TotalKeys != 0 {
		t.Errorf("expected empty report")
	}
}
