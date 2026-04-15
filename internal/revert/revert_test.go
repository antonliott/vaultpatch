package revert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/revert"
)

type mockReader struct {
	versions map[int]map[string]string
	readErr  error
	writeErr error
	written  map[string]string
}

func (m *mockReader) ReadVersion(_ context.Context, _ string, version int) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	d, ok := m.versions[version]
	if !ok {
		return nil, errors.New("version not found")
	}
	return d, nil
}

func (m *mockReader) Write(_ context.Context, _ string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	mr := &mockReader{versions: map[int]map[string]string{2: {"key": "old"}}}
	rv := revert.NewReverter(mr)
	res := rv.Apply(context.Background(), revert.Options{Path: "secret/foo", Version: 2, DryRun: true})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if mr.written != nil {
		t.Error("dry run must not write")
	}
}

func TestApply_Success(t *testing.T) {
	mr := &mockReader{versions: map[int]map[string]string{3: {"db": "pass"}}}
	rv := revert.NewReverter(mr)
	res := rv.Apply(context.Background(), revert.Options{Path: "secret/db", Version: 3})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if mr.written["db"] != "pass" {
		t.Errorf("expected written data to match version 3")
	}
}

func TestApply_ReadError(t *testing.T) {
	mr := &mockReader{readErr: errors.New("vault unavailable")}
	rv := revert.NewReverter(mr)
	res := rv.Apply(context.Background(), revert.Options{Path: "secret/x", Version: 1})
	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_InvalidVersion(t *testing.T) {
	mr := &mockReader{}
	rv := revert.NewReverter(mr)
	res := rv.Apply(context.Background(), revert.Options{Path: "secret/x", Version: 0})
	if res.Err == nil {
		t.Fatal("expected error for version 0")
	}
}

func TestApply_WriteError(t *testing.T) {
	mr := &mockReader{
		versions: map[int]map[string]string{1: {"a": "b"}},
		writeErr: errors.New("permission denied"),
	}
	rv := revert.NewReverter(mr)
	res := rv.Apply(context.Background(), revert.Options{Path: "secret/y", Version: 1})
	if res.Err == nil {
		t.Fatal("expected write error")
	}
}
