package baseline_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/baseline"
)

type mockReader struct {
	listFn func(ctx context.Context, path string) ([]string, error)
	readFn func(ctx context.Context, path string) (map[string]string, error)
}

func (m *mockReader) List(ctx context.Context, path string) ([]string, error) {
	return m.listFn(ctx, path)
}
func (m *mockReader) Read(ctx context.Context, path string) (map[string]string, error) {
	return m.readFn(ctx, path)
}

func TestCapture_Success(t *testing.T) {
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"db", "api"}, nil
		},
		readFn: func(_ context.Context, path string) (map[string]string, error) {
			if path == "secret/db" {
				return map[string]string{"password": "s3cr3t"}, nil
			}
			return map[string]string{"token": "abc"}, nil
		},
	}
	b, err := baseline.Capture(context.Background(), r, "ns1", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Namespace != "ns1" {
		t.Errorf("expected namespace ns1, got %s", b.Namespace)
	}
	if len(b.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(b.Secrets))
	}
	if b.CapturedAt.IsZero() {
		t.Error("expected non-zero CapturedAt")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	b := &baseline.Baseline{
		CapturedAt: time.Now().UTC().Truncate(time.Second),
		Namespace:  "dev",
		Secrets:    map[string]map[string]string{"svc": {"key": "val"}},
	}
	tmp := filepath.Join(t.TempDir(), "baseline.json")
	if err := baseline.Save(b, tmp); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := baseline.Load(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Namespace != b.Namespace {
		t.Errorf("namespace mismatch: %s != %s", loaded.Namespace, b.Namespace)
	}
	if loaded.Secrets["svc"]["key"] != "val" {
		t.Errorf("secret value mismatch")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(tmp, []byte("not-json"), 0o600)
	_, err := baseline.Load(tmp)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCompare_DetectsDrift(t *testing.T) {
	b := &baseline.Baseline{
		Secrets: map[string]map[string]string{
			"svc": {"token": "old", "host": "localhost"},
		},
	}
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"svc"}, nil
		},
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"token": "new", "extra": "added"}, nil
		},
	}
	deltas, err := baseline.Compare(context.Background(), r, b, "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	statuses := map[string]string{}
	for _, d := range deltas {
		statuses[d.Key] = d.Status
	}
	if statuses["token"] != "changed" {
		t.Errorf("expected token=changed, got %s", statuses["token"])
	}
	if statuses["host"] != "removed" {
		t.Errorf("expected host=removed, got %s", statuses["host"])
	}
	if statuses["extra"] != "added" {
		t.Errorf("expected extra=added, got %s", statuses["extra"])
	}
}

func TestCompare_NoDrift(t *testing.T) {
	b := &baseline.Baseline{
		Secrets: map[string]map[string]string{
			"svc": {"token": "abc"},
		},
	}
	r := &mockReader{
		listFn: func(_ context.Context, _ string) ([]string, error) { return []string{"svc"}, nil },
		readFn: func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"token": "abc"}, nil
		},
	}
	deltas, err := baseline.Compare(context.Background(), r, b, "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deltas) != 0 {
		t.Errorf("expected no deltas, got %d", len(deltas))
	}
}

// ensure Baseline is JSON-serialisable (compile-time check via json.Marshal)
func TestBaseline_JSONRoundTrip(t *testing.T) {
	b := &baseline.Baseline{
		CapturedAt: time.Now().UTC(),
		Namespace:  "prod",
		Secrets:    map[string]map[string]string{"x": {"k": "v"}},
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var b2 baseline.Baseline
	if err := json.Unmarshal(data, &b2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if b2.Namespace != "prod" {
		t.Errorf("namespace mismatch")
	}
}
