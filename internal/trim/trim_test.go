package trim_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/trim"
)

type mockClient struct {
	keys     []string
	meta     map[string]map[string]interface{}
	deleted  []string
	listErr  error
	readErr  map[string]error
	delErr   map[string]error
}

func (m *mockClient) ListSecrets(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.keys, nil
}

func (m *mockClient) ReadSecretMetadata(_ context.Context, path string) (map[string]interface{}, error) {
	if err, ok := m.readErr[path]; ok {
		return nil, err
	}
	if meta, ok := m.meta[path]; ok {
		return meta, nil
	}
	return map[string]interface{}{}, nil
}

func (m *mockClient) DeleteSecret(_ context.Context, path string) error {
	if err, ok := m.delErr[path]; ok {
		return err
	}
	m.deleted = append(m.deleted, path)
	return nil
}

func ago(d time.Duration) string {
	return time.Now().UTC().Add(-d).Format(time.RFC3339Nano)
}

func TestApply_DryRun(t *testing.T) {
	client := &mockClient{
		keys: []string{"old-key"},
		meta: map[string]map[string]interface{}{
			"secret/old-key": {"created_time": ago(48 * time.Hour)},
		},
	}
	tr := trim.NewTrimmer(client, 24*time.Hour, true)
	res, err := tr.Apply(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Deleted) != 1 {
		t.Errorf("expected 1 deleted, got %d", len(res.Deleted))
	}
	if len(client.deleted) != 0 {
		t.Error("dry-run should not call DeleteSecret")
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestApply_Success(t *testing.T) {
	client := &mockClient{
		keys: []string{"old", "new"},
		meta: map[string]map[string]interface{}{
			"secret/old": {"created_time": ago(72 * time.Hour)},
			"secret/new": {"created_time": ago(1 * time.Hour)},
		},
	}
	tr := trim.NewTrimmer(client, 24*time.Hour, false)
	res, err := tr.Apply(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Deleted) != 1 || res.Deleted[0] != "secret/old" {
		t.Errorf("expected [secret/old] deleted, got %v", res.Deleted)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(res.Skipped))
	}
}

func TestApply_ListError(t *testing.T) {
	client := &mockClient{listErr: errors.New("permission denied")}
	tr := trim.NewTrimmer(client, time.Hour, false)
	_, err := tr.Apply(context.Background(), "secret")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApply_ReadError(t *testing.T) {
	client := &mockClient{
		keys:    []string{"key1"},
		readErr: map[string]error{"secret/key1": fmt.Errorf("read failed")},
	}
	tr := trim.NewTrimmer(client, time.Hour, false)
	res, err := tr.Apply(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if !res.HasErrors() {
		t.Error("expected errors in result")
	}
}

func TestResult_Summary(t *testing.T) {
	res := &trim.Result{Deleted: []string{"a", "b"}, Skipped: []string{"c"}, DryRun: false}
	s := res.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}
