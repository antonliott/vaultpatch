package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/vaultpatch/internal/snapshot"
)

func TestNew_SetsFields(t *testing.T) {
	secrets := map[string]string{"key": "value"}
	s := snapshot.New("ns1", secrets)

	if s.Namespace != "ns1" {
		t.Errorf("expected namespace ns1, got %s", s.Namespace)
	}
	if s.Secrets["key"] != "value" {
		t.Errorf("expected secrets to contain key=value")
	}
	if s.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	orig := snapshot.New("prod", map[string]string{"db/pass": "secret123"})
	if err := snapshot.Save(orig, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Namespace != orig.Namespace {
		t.Errorf("namespace mismatch: got %s, want %s", loaded.Namespace, orig.Namespace)
	}
	if loaded.Secrets["db/pass"] != "secret123" {
		t.Errorf("secret mismatch")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	s := snapshot.New("ns", map[string]string{})
	err := snapshot.Save(s, "/nonexistent/dir/snap.json")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not-json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := snapshot.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
