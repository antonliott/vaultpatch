package stamp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/stamp"
)

type mockRW struct {
	data    map[string]map[string]string
	readErr error
	writeErr error
	wrote   map[string]map[string]string
}

func (m *mockRW) Read(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.data[path], nil
}

func (m *mockRW) Write(_ context.Context, path string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.wrote == nil {
		m.wrote = make(map[string]map[string]string)
	}
	m.wrote[path] = data
	return nil
}

func TestApply_StampsPath(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"secret/app": {"key": "val"}}}
	s := stamp.New(rw, stamp.Options{})
	res := s.Apply(context.Background(), []string{"secret/app"})
	if len(res) != 1 || res[0].Error != nil {
		t.Fatalf("unexpected error: %v", res[0].Error)
	}
	if _, ok := rw.wrote["secret/app"][stamp.DefaultKey]; !ok {
		t.Fatal("stamp key not written")
	}
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"secret/app": {}}}
	s := stamp.New(rw, stamp.Options{DryRun: true})
	res := s.Apply(context.Background(), []string{"secret/app"})
	if res[0].Error != nil {
		t.Fatal(res[0].Error)
	}
	if len(rw.wrote) != 0 {
		t.Fatal("expected no writes in dry-run")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("permission denied")}
	s := stamp.New(rw, stamp.Options{})
	res := s.Apply(context.Background(), []string{"secret/app"})
	if res[0].Error == nil {
		t.Fatal("expected error")
	}
}

func TestApply_WithActor(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{"secret/app": {}}}
	s := stamp.New(rw, stamp.Options{Actor: "ci-bot"})
	s.Apply(context.Background(), []string{"secret/app"})
	v := rw.wrote["secret/app"][stamp.DefaultKey]
	if v == "" {
		t.Fatal("stamp not written")
	}
	if len(v) < 5 {
		t.Fatalf("unexpected stamp value: %s", v)
	}
}
