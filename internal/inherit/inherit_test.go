package inherit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/inherit"
)

type mockRW struct {
	data    map[string]map[string]string
	writes  map[string]map[string]string
	readErr map[string]error
	writeErr map[string]error
}

func newMockRW() *mockRW {
	return &mockRW{
		data:     make(map[string]map[string]string),
		writes:   make(map[string]map[string]string),
		readErr:  make(map[string]error),
		writeErr: make(map[string]error),
	}
}

func (m *mockRW) ReadSecrets(_ context.Context, path string) (map[string]string, error) {
	if err, ok := m.readErr[path]; ok {
		return nil, err
	}
	if d, ok := m.data[path]; ok {
		copy := make(map[string]string, len(d))
		for k, v := range d {
			copy[k] = v
		}
		return copy, nil
	}
	return map[string]string{}, nil
}

func (m *mockRW) WriteSecrets(_ context.Context, path string, data map[string]string) error {
	if err, ok := m.writeErr[path]; ok {
		return err
	}
	m.writes[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"db_pass": "s3cr3t"}
	rw.data["secret/child"] = map[string]string{"app_key": "abc"}

	inheritor := inherit.NewInheritor(rw, inherit.Options{DryRun: true})
	res, err := inheritor.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Children[0].Applied != 1 {
		t.Errorf("expected 1 applied, got %d", res.Children[0].Applied)
	}
	if _, wrote := rw.writes["secret/child"]; wrote {
		t.Error("dry run should not write")
	}
}

func TestApply_Success_NoOverwrite(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"key": "parent_val", "shared": "from_parent"}
	rw.data["secret/child"] = map[string]string{"shared": "child_val"}

	inheritor := inherit.NewInheritor(rw, inherit.Options{})
	res, err := inheritor.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cr := res.Children[0]
	if cr.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", cr.Applied)
	}
	if cr.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", cr.Skipped)
	}
	if rw.writes["secret/child"]["shared"] != "child_val" {
		t.Error("existing child value should not be overwritten")
	}
}

func TestApply_Success_Overwrite(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"key": "parent_val"}
	rw.data["secret/child"] = map[string]string{"key": "old_val"}

	inheritor := inherit.NewInheritor(rw, inherit.Options{Overwrite: true})
	_, err := inheritor.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rw.writes["secret/child"]["key"] != "parent_val" {
		t.Error("overwrite mode should replace existing child value")
	}
}

func TestApply_ParentReadError(t *testing.T) {
	rw := newMockRW()
	rw.readErr["secret/parent"] = errors.New("permission denied")

	inheritor := inherit.NewInheritor(rw, inherit.Options{})
	_, err := inheritor.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_ChildReadError(t *testing.T) {
	rw := newMockRW()
	rw.data["secret/parent"] = map[string]string{"key": "val"}
	rw.readErr["secret/child"] = errors.New("not found")

	inheritor := inherit.NewInheritor(rw, inherit.Options{})
	res, err := inheritor.Apply(context.Background(), "secret/parent", []string{"secret/child"})
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if res.Children[0].Err == nil {
		t.Error("expected child error, got nil")
	}
}
