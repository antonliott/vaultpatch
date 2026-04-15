package dedupe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/dedupe"
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

func TestFind_NoDuplicates(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"a", "b"}, nil
		},
		readFn: func(_ context.Context, p string) (map[string]string, error) {
			if p == "a" {
				return map[string]string{"key": "alpha"}, nil
			}
			return map[string]string{"key": "beta"}, nil
		},
	}
	res, err := dedupe.NewFinder(r).Find(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasDuplicates() {
		t.Fatalf("expected no duplicates, got %d", len(res.Duplicates))
	}
	if got := res.Summary(); got != "no duplicate secret values found" {
		t.Errorf("unexpected summary: %q", got)
	}
}

func TestFind_DetectsDuplicate(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"svc/a", "svc/b", "svc/c"}, nil
		},
		readFn: func(_ context.Context, p string) (map[string]string, error) {
			switch p {
			case "svc/a":
				return map[string]string{"db_pass": "s3cr3t"}, nil
			case "svc/b":
				return map[string]string{"db_pass": "s3cr3t"}, nil // duplicate
			default:
				return map[string]string{"db_pass": "unique"}, nil
			}
		},
	}
	res, err := dedupe.NewFinder(r).Find(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasDuplicates() {
		t.Fatal("expected duplicates but found none")
	}
	if len(res.Duplicates) != 1 {
		t.Fatalf("expected 1 duplicate groupres.Duplicates))
	}
	d := res.Duplicates[0]
	if d.Key != "db_pass" {
		t.Errorf("expected key db_pass, got %q", d.Key)
	}
	if len(d.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(d.Paths))
	}
}

func TestFind_ListError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return nil, errors.New("permission denied")
		},
		readFn: nil,
	}
	_, err := dedupe.NewFinder(r).Find(context.Background(), "secret/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFind_ReadError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"x"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, errors.New("read failed")
		},
	}
	_, err := dedupe.NewFinder(r).Find(context.Background(), "secret/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
