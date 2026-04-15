package mirror_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/mirror"
)

type mockStore struct {
	data    map[string]map[string]string
	listErr error
	readErr map[string]error
	written map[string]map[string]string
}

func (m *mockStore) List(_ context.Context, path string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var keys []string
	for k := range m.data {
		if len(k) > len(path) && k[:len(path)] == path {
			keys = append(keys, k[len(path)+1:])
		}
	}
	return keys, nil
}

func (m *mockStore) Read(_ context.Context, path string) (map[string]string, error) {
	if err, ok := m.readErr[path]; ok {
		return nil, err
	}
	return m.data[path], nil
}

func (m *mockStore) Write(_ context.Context, path string, data map[string]string) error {
	if m.written == nil {
		m.written = make(map[string]map[string]string)
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	store := &mockStore{
		data: map[string]map[string]string{
			"src/db": {"password": "secret"},
		},
	}
	m := mirror.NewMirrorer(store, store, true, true)
	res, err := m.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Mirrored) != 1 {
		t.Errorf("expected 1 mirrored, got %d", len(res.Mirrored))
	}
	if len(store.written) != 0 {
		t.Error("dry-run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	store := &mockStore{
		data: map[string]map[string]string{
			"src/api": {"key": "val"},
		},
	}
	m := mirror.NewMirrorer(store, store, true, false)
	res, err := m.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Mirrored) != 1 || res.Mirrored[0] != "dst/api" {
		t.Errorf("unexpected mirrored: %v", res.Mirrored)
	}
	if _, ok := store.written["dst/api"]; !ok {
		t.Error("expected dst/api to be written")
	}
}

func TestApply_SkipsExistingWithoutOverwrite(t *testing.T) {
	store := &mockStore{
		data: map[string]map[string]string{
			"src/cfg": {"a": "1"},
			"dst/cfg": {"a": "old"},
		},
	}
	m := mirror.NewMirrorer(store, store, false, false)
	res, err := m.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(res.Skipped))
	}
}

func TestApply_ListError(t *testing.T) {
	store := &mockStore{listErr: errors.New("permission denied")}
	m := mirror.NewMirrorer(store, store, true, false)
	_, err := m.Apply(context.Background(), "src", "dst")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResult_Summary(t *testing.T) {
	r := &mirror.Result{Mirrored: []string{"a", "b"}, Skipped: []string{"c"}, DryRun: false}
	s := r.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}
