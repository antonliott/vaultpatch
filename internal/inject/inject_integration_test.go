package inject_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultpatch/internal/inject"
)

// staticStore is an in-memory Vault reader for integration tests.
type staticStore struct {
	paths map[string]map[string]string
}

func (s *staticStore) Read(_ context.Context, path string) (map[string]string, error) {
	v, ok := s.paths[path]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func TestInject_MultiplePathsSingleRead(t *testing.T) {
	store := &staticStore{
		paths: map[string]map[string]string{
			"secret/app": {"api_key": "abc123", "region": "us-east-1"},
			"secret/db":  {"host": "db.internal", "port": "5432"},
		},
	}

	input := map[string]string{
		"API_KEY":  "vault://secret/app#api_key",
		"REGION":   "vault://secret/app#region",
		"DB_HOST":  "vault://secret/db#host",
		"DB_PORT":  "vault://secret/db#port",
		"STATIC":   "plainvalue",
	}

	inj := inject.New(store, inject.Options{})
	out, res := inj.Apply(context.Background(), input)

	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if res.Injected != 4 {
		t.Errorf("expected 4 injected, got %d", res.Injected)
	}
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", res.Skipped)
	}
	assertEq(t, out["API_KEY"], "abc123")
	assertEq(t, out["REGION"], "us-east-1")
	assertEq(t, out["DB_HOST"], "db.internal")
	assertEq(t, out["DB_PORT"], "5432")
	assertEq(t, out["STATIC"], "plainvalue")
}

func TestInject_CustomPrefix(t *testing.T) {
	store := &staticStore{
		paths: map[string]map[string]string{
			"kv/svc": {"token": "tok-xyz"},
		},
	}
	input := map[string]string{"SVC_TOKEN": "secret://kv/svc#token"}
	inj := inject.New(store, inject.Options{Prefix: "secret://"})
	out, res := inj.Apply(context.Background(), input)
	if res.HasErrors() {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	assertEq(t, out["SVC_TOKEN"], "tok-xyz")
}

func assertEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
