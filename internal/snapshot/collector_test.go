package snapshot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/snapshot"
)

type mockReader struct {
	listFn func(ctx context.Context, path string) ([]string, error)
	readFn func(ctx context.Context, path string) (map[string]string, error)
}

func (m *mockReader) ListSecrets(ctx context.Context, path string) ([]string, error) {
	return m.listFn(ctx, path)
}

func (m *mockReader) ReadSecret(ctx context.Context, path string) (map[string]string, error) {
	return m.readFn(ctx, path)
}

func TestCollect_Success(t *testing.T) {
	reader := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"secret/app"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"password": "hunter2"}, nil
		},
	}

	c := snapshot.NewCollector("dev", reader)
	snap, err := c.Collect(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Namespace != "dev" {
		t.Errorf("expected namespace dev, got %s", snap.Namespace)
	}
	if snap.Secrets["secret/app/password"] != "hunter2" {
		t.Errorf("expected secret value hunter2")
	}
}

func TestCollect_ListError(t *testing.T) {
	reader := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return nil, errors.New("permission denied")
		},
		readFn: nil,
	}
	c := snapshot.NewCollector("dev", reader)
	_, err := c.Collect(context.Background(), "secret")
	if err == nil {
		t.Error("expected error from list, got nil")
	}
}

func TestCollect_ReadError(t *testing.T) {
	reader := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"secret/app"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, errors.New("read failed")
		},
	}
	c := snapshot.NewCollector("dev", reader)
	_, err := c.Collect(context.Background(), "secret")
	if err == nil {
		t.Error("expected error from read, got nil")
	}
}

func TestCollect_EmptyList(t *testing.T) {
	reader := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{}, nil
		},
		readFn: nil,
	}
	c := snapshot.NewCollector("dev", reader)
	snap, err := c.Collect(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Secrets) != 0 {
		t.Errorf("expected empty secrets map, got %d entries", len(snap.Secrets))
	}
}
