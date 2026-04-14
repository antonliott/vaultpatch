package scope_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/scope"
)

type mockReader struct {
	paths []string
	err   error
}

func (m *mockReader) List(_ context.Context, _, _ string) ([]string, error) {
	return m.paths, m.err
}

func TestResolve_NoFilter(t *testing.T) {
	r := &mockReader{paths: []string{"db/password", "api/key", "tls/cert"}}
	s := scope.NewScope(r, "secret")

	got, err := s.Resolve(context.Background(), scope.Filter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(got))
	}
}

func TestResolve_ExcludesPaths(t *testing.T) {
	r := &mockReader{paths: []string{"db/password", "api/key", "tls/cert"}}
	s := scope.NewScope(r, "secret")

	got, err := s.Resolve(context.Background(), scope.Filter{Exclude: []string{"api/key"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(got))
	}
	for _, p := range got {
		if p == "api/key" {
			t.Error("excluded path should not appear in results")
		}
	}
}

func TestResolve_EmptyResult(t *testing.T) {
	r := &mockReader{paths: []string{}}
	s := scope.NewScope(r, "secret")

	got, err := s.Resolve(context.Background(), scope.Filter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(got))
	}
}

func TestResolve_ReaderError(t *testing.T) {
	r := &mockReader{err: errors.New("vault unavailable")}
	s := scope.NewScope(r, "secret")

	_, err := s.Resolve(context.Background(), scope.Filter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, r.err) {
		t.Errorf("error chain should wrap original: %v", err)
	}
}

func TestResolve_MultipleExcludes(t *testing.T) {
	r := &mockReader{paths: []string{"a", "b", "c", "d"}}
	s := scope.NewScope(r, "kv")

	got, err := s.Resolve(context.Background(), scope.Filter{Exclude: []string{"b", "d"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(got), got)
	}
}
