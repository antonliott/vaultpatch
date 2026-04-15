package inject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/inject"
)

type mockReader struct {
	data map[string]map[string]string
	err  error
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	v, ok := m.data[path]
	if !ok {
		return nil, errors.New("path not found")
	}
	return v, nil
}

func TestApply_NoReferences(t *testing.T) {
	inj := inject.New(&mockReader{}, inject.Options{})
	input := map[string]string{"FOO": "bar", "BAZ": "qux"}
	out, res := inj.Apply(context.Background(), input)
	if res.Injected != 0 || res.Skipped != 2 || res.HasErrors() {
		t.Fatalf("unexpected result: %+v", res)
	}
	if out["FOO"] != "bar" {
		t.Errorf("expected bar, got %s", out["FOO"])
	}
}

func TestApply_InjectsValue(t *testing.T) {
	r := &mockReader{data: map[string]map[string]string{
		"secret/db": {"password": "s3cr3t"},
	}}
	inj := inject.New(r, inject.Options{})
	input := map[string]string{"DB_PASS": "vault://secret/db#password"}
	out, res := inj.Apply(context.Background(), input)
	if res.Injected != 1 || res.HasErrors() {
		t.Fatalf("unexpected result: %+v", res)
	}
	if out["DB_PASS"] != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %s", out["DB_PASS"])
	}
}

func TestApply_DryRun(t *testing.T) {
	inj := inject.New(&mockReader{}, inject.Options{DryRun: true})
	input := map[string]string{"TOKEN": "vault://secret/app#token"}
	out, res := inj.Apply(context.Background(), input)
	if res.Injected != 1 || res.HasErrors() {
		t.Fatalf("unexpected result: %+v", res)
	}
	if out["TOKEN"] != "<injected:secret/app#token>" {
		t.Errorf("unexpected dry-run value: %s", out["TOKEN"])
	}
}

func TestApply_MissingField(t *testing.T) {
	r := &mockReader{data: map[string]map[string]string{
		"secret/app": {"other": "val"},
	}}
	inj := inject.New(r, inject.Options{})
	input := map[string]string{"X": "vault://secret/app#missing"}
	_, res := inj.Apply(context.Background(), input)
	if !res.HasErrors() {
		t.Fatal("expected error for missing field")
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{err: errors.New("connection refused")}
	inj := inject.New(r, inject.Options{})
	input := map[string]string{"X": "vault://secret/db#pass"}
	_, res := inj.Apply(context.Background(), input)
	if !res.HasErrors() {
		t.Fatal("expected error on read failure")
	}
}

func TestApply_InvalidReference(t *testing.T) {
	inj := inject.New(&mockReader{}, inject.Options{})
	input := map[string]string{"X": "vault://secret/db"} // missing #field
	_, res := inj.Apply(context.Background(), input)
	if !res.HasErrors() {
		t.Fatal("expected error for invalid reference")
	}
}

func TestApply_CachesReads(t *testing.T) {
	calls := 0
	r := &mockReader{data: map[string]map[string]string{
		"secret/db": {"user": "admin", "pass": "secret"},
	}}
	_ = calls
	inj := inject.New(r, inject.Options{})
	input := map[string]string{
		"A": "vault://secret/db#user",
		"B": "vault://secret/db#pass",
	}
	out, res := inj.Apply(context.Background(), input)
	if res.Injected != 2 || res.HasErrors() {
		t.Fatalf("unexpected result: %+v", res)
	}
	if out["A"] != "admin" || out["B"] != "secret" {
		t.Errorf("unexpected values: %v", out)
	}
}
