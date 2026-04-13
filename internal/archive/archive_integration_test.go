package archive_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/your-org/vaultpatch/internal/archive"
)

// inMemStore satisfies both Reader and Writer for integration-style tests.
type inMemStore struct {
	store map[string]map[string]interface{}
}

func newInMemStore(initial map[string]map[string]interface{}) *inMemStore {
	if initial == nil {
		initial = make(map[string]map[string]interface{})
	}
	return &inMemStore{store: initial}
}

func (s *inMemStore) Read(_ context.Context, path string) (map[string]interface{}, error) {
	return s.store[path], nil
}

func (s *inMemStore) List(_ context.Context, prefix string) ([]string, error) {
	var keys []string
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *inMemStore) Write(_ context.Context, path string, data map[string]interface{}) error {
	s.store[path] = data
	return nil
}

func TestArchive_MultipleSecrets(t *testing.T) {
	paths := []string{"prod/db", "prod/api", "prod/cache"}
	initial := make(map[string]map[string]interface{})
	for i, p := range paths {
		initial[p] = map[string]interface{}{"key": fmt.Sprintf("value%d", i)}
	}
	store := newInMemStore(initial)
	a := archive.NewArchiver(store, store, "_archive", false)

	res, err := a.Apply(context.Background(), paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Archived) != 3 {
		t.Errorf("expected 3 archived, got %d", len(res.Archived))
	}
	for _, p := range paths {
		dest := "_archive/" + p
		if _, ok := store.store[dest]; !ok {
			t.Errorf("expected archived entry at %s", dest)
		}
	}
}

func TestArchive_PreservesOriginalData(t *testing.T) {
	store := newInMemStore(map[string]map[string]interface{}{
		"prod/db": {"password": "hunter2", "user": "admin"},
	})
	a := archive.NewArchiver(store, store, "_archive", false)
	_, err := a.Apply(context.Background(), []string{"prod/db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	archived := store.store["_archive/prod/db"]
	if archived["password"] != "hunter2" || archived["user"] != "admin" {
		t.Error("original secret data not preserved in archive")
	}
}
