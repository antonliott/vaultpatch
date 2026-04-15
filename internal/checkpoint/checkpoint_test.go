package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/youorg/vaultpatch/internal/checkpoint"
)

func TestNew_SetsFields(t *testing.T) {
	secrets := map[string]checkpoint.Secret{
		"secret/foo": {Data: map[string]string{"key": "val"}},
	}
	cp := checkpoint.New("v1", "prod", secrets)
	if cp.Label != "v1" {
		t.Errorf("expected label v1, got %s", cp.Label)
	}
	if cp.Namespace != "prod" {
		t.Errorf("expected namespace prod, got %s", cp.Namespace)
	}
	if len(cp.Paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(cp.Paths))
	}
	if cp.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	secrets := map[string]checkpoint.Secret{
		"secret/db": {Data: map[string]string{"pass": "hunter2"}},
	}
	cp := checkpoint.New("release-1.2", "staging", secrets)
	cp.CreatedAt = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	tmp := filepath.Join(t.TempDir(), "cp.json")
	if err := checkpoint.Save(cp, tmp); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := checkpoint.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Label != cp.Label {
		t.Errorf("label mismatch: got %s", loaded.Label)
	}
	if loaded.Secrets["secret/db"].Data["pass"] != "hunter2" {
		t.Error("secret data mismatch")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := checkpoint.Load("/nonexistent/checkpoint.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	cp := checkpoint.New("x", "ns", nil)
	err := checkpoint.Save(cp, "/no/such/dir/cp.json")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(tmp, []byte("not-json{"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := checkpoint.Load(tmp)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
