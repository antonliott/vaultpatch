package digest_test

import (
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/digest"
)

// staticReader is a simple in-memory Reader for tests.
type staticReader struct {
	keys   []string
	values map[string]map[string]string
	listErr error
	readErr map[string]error
}

func (s *staticReader) List(_ string) ([]string, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.keys, nil
}

func (s *staticReader) Read(path string) (map[string]string, error) {
	if err, ok := s.readErr[path]; ok {
		return nil, err
	}
	return s.values[path], nil
}

func TestCompute_Success(t *testing.T) {
	r := &staticReader{
		keys: []string{"db"},
		values: map[string]map[string]string{
			"secret/db": {"password": "s3cr3t"},
		},
	}
	res, err := digest.Compute(r, "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(res))
	}
	entry, ok := res["secret/db"]
	if !ok {
		t.Fatal("missing entry for secret/db")
	}
	if entry.Digest == "" {
		t.Error("digest should not be empty")
	}
}

func TestCompute_ListError(t *testing.T) {
	r := &staticReader{listErr: errors.New("forbidden")}
	_, err := digest.Compute(r, "secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCompute_ReadError(t *testing.T) {
	r := &staticReader{
		keys: []string{"db"},
		readErr: map[string]error{"secret/db": errors.New("not found")},
		values:  map[string]map[string]string{},
	}
	_, err := digest.Compute(r, "secret")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChanged_NoDiff(t *testing.T) {
	a := digest.Result{"secret/db": {Path: "secret/db", Digest: "abc"}}
	b := digest.Result{"secret/db": {Path: "secret/db", Digest: "abc"}}
	if got := digest.Changed(a, b); len(got) != 0 {
		t.Fatalf("expected no changes, got %v", got)
	}
}

func TestChanged_Modified(t *testing.T) {
	a := digest.Result{"secret/db": {Path: "secret/db", Digest: "abc"}}
	b := digest.Result{"secret/db": {Path: "secret/db", Digest: "xyz"}}
	got := digest.Changed(a, b)
	if len(got) != 1 || got[0] != "secret/db" {
		t.Fatalf("expected [secret/db], got %v", got)
	}
}

func TestChanged_Added(t *testing.T) {
	a := digest.Result{}
	b := digest.Result{"secret/new": {Path: "secret/new", Digest: "111"}}
	got := digest.Changed(a, b)
	if len(got) != 1 || got[0] != "secret/new" {
		t.Fatalf("expected [secret/new], got %v", got)
	}
}

func TestChanged_Removed(t *testing.T) {
	a := digest.Result{"secret/old": {Path: "secret/old", Digest: "999"}}
	b := digest.Result{}
	got := digest.Changed(a, b)
	if len(got) != 1 || got[0] != "secret/old" {
		t.Fatalf("expected [secret/old], got %v", got)
	}
}

func TestDigest_IsStable(t *testing.T) {
	r1 := &staticReader{
		keys:   []string{"api"},
		values: map[string]map[string]string{"secret/api": {"key": "val", "token": "tok"}},
	}
	r2 := &staticReader{
		keys:   []string{"api"},
		values: map[string]map[string]string{"secret/api": {"token": "tok", "key": "val"}},
	}
	res1, _ := digest.Compute(r1, "secret")
	res2, _ := digest.Compute(r2, "secret")
	if digest.Changed(res1, res2) != nil {
		t.Error("digest should be stable regardless of map iteration order")
	}
}
