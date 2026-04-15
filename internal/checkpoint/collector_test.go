package checkpoint_test

import (
	"context"
	"errors"
	"testing"

	"github.com/youorg/vaultpatch/internal/checkpoint"
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
			return []string{"db"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"pass": "s3cr3t"}, nil
		},
	}
	col := checkpoint.NewCollector(r, "prod")
	secrets, err := col.Collect(context.Background(), []string{"secret/app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
	if secrets["secret/app/db"].Data["pass"] != "s3cr3t" {
		t.Error("unexpected secret value")
	}
}

func TestCollect_ListError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return nil, errors.New("permission denied")
		},
		readFn: nil,
	}
	col := checkpoint.NewCollector(r, "prod")
	_, err := col.Collect(context.Background(), []string{"secret/app"})
	if err == nil {
		t.Error("expected error from List")
	}
}

func TestCollect_ReadError(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"key"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, errors.New("read failed")
		},
	}
	col := checkpoint.NewCollector(r, "prod")
	_, err := col.Collect(context.Background(), []string{"secret/app"})
	if err == nil {
		t.Error("expected error from Read")
	}
}
