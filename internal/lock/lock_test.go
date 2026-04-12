package lock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/lock"
)

type mockWriter struct {
	store  map[string]map[string]interface{}
	writeErr error
	deleteErr error
}

func newMockWriter() *mockWriter {
	return &mockWriter{store: make(map[string]map[string]interface{})}
}

func (m *mockWriter) Write(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.store[path] = data
	return nil
}

func (m *mockWriter) Delete(_ context.Context, path string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.store, path)
	return nil
}

func (m *mockWriter) Read(_ context.Context, path string) (map[string]interface{}, error) {
	v, ok := m.store[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func TestAcquire_Success(t *testing.T) {
	w := newMockWriter()
	l := lock.NewLock(w, "locks/mypath", "worker-1", 30*time.Second)
	if err := l.Acquire(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAcquire_AlreadyHeld(t *testing.T) {
	w := newMockWriter()
	l1 := lock.NewLock(w, "locks/mypath", "worker-1", 30*time.Second)
	_ = l1.Acquire(context.Background())

	l2 := lock.NewLock(w, "locks/mypath", "worker-2", 30*time.Second)
	if err := l2.Acquire(context.Background()); err == nil {
		t.Fatal("expected error for held lock")
	}
}

func TestAcquire_ExpiredLock(t *testing.T) {
	w := newMockWriter()
	l1 := lock.NewLock(w, "locks/mypath", "worker-1", 1*time.Millisecond)
	_ = l1.Acquire(context.Background())
	time.Sleep(5 * time.Millisecond)

	l2 := lock.NewLock(w, "locks/mypath", "worker-2", 30*time.Second)
	if err := l2.Acquire(context.Background()); err != nil {
		t.Fatalf("expected expired lock to be overridden, got %v", err)
	}
}

func TestRelease_Success(t *testing.T) {
	w := newMockWriter()
	l := lock.NewLock(w, "locks/mypath", "worker-1", 30*time.Second)
	_ = l.Acquire(context.Background())
	if err := l.Release(context.Background()); err != nil {
		t.Fatalf("expected no error on release, got %v", err)
	}
}

func TestRelease_WriteError(t *testing.T) {
	w := newMockWriter()
	w.deleteErr = errors.New("vault unavailable")
	l := lock.NewLock(w, "locks/mypath", "worker-1", 30*time.Second)
	if err := l.Release(context.Background()); err == nil {
		t.Fatal("expected error on release failure")
	}
}

func TestIsExpired(t *testing.T) {
	w := newMockWriter()
	l := lock.NewLock(w, "locks/mypath", "worker-1", 1*time.Millisecond)
	_ = l.Acquire(context.Background())
	time.Sleep(5 * time.Millisecond)
	if !l.IsExpired() {
		t.Fatal("expected lock to be expired")
	}
}
