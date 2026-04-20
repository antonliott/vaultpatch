package pivot_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/pivot"
)

// mockReader is a simple in-memory Reader.
type mockReader struct {
	paths  []string
	secrets map[string]map[string]string
	listErr error
	readErr map[string]error
}

func (m *mockReader) List(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.paths, nil
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]string, error) {
	if err, ok := m.readErr[path]; ok {
		return nil, err
	}
	return m.secrets[path], nil
}

func TestApply_NoPaths(t *testing.T) {
	r := &mockReader{}
	p := pivot.NewPivoter(r)
	res, err := p.Apply(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected empty result, got %d keys", len(res))
	}
}

func TestApply_Success(t *testing.T) {
	r := &mockReader{
		paths: []string{"secret/app1", "secret/app2"},
		secrets: map[string]map[string]string{
			"secret/app1": {"DB_HOST": "host1", "API_KEY": "k1"},
			"secret/app2": {"DB_HOST": "host2"},
		},
	}
	p := pivot.NewPivoter(r)
	res, err := p.Apply(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["DB_HOST"]["secret/app1"] != "host1" {
		t.Errorf("expected host1, got %q", res["DB_HOST"]["secret/app1"])
	}
	if res["DB_HOST"]["secret/app2"] != "host2" {
		t.Errorf("expected host2, got %q", res["DB_HOST"]["secret/app2"])
	}
	if _, ok := res["API_KEY"]; !ok {
		t.Error("expected API_KEY in result")
	}
}

func TestApply_ListError(t *testing.T) {
	r := &mockReader{listErr: errors.New("permission denied")}
	p := pivot.NewPivoter(r)
	_, err := p.Apply(context.Background(), "secret/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{
		paths:   []string{"secret/app1"},
		readErr: map[string]error{"secret/app1": errors.New("read fail")},
	}
	p := pivot.NewPivoter(r)
	_, err := p.Apply(context.Background(), "secret/")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFormat_Empty(t *testing.T) {
	out := pivot.Format(pivot.Result{})
	if !strings.Contains(out, "no keys") {
		t.Errorf("expected 'no keys' message, got %q", out)
	}
}

func TestFormat_NonEmpty(t *testing.T) {
	res := pivot.Result{
		"DB_HOST": {"secret/app1": "host1", "secret/app2": "host2"},
	}
	out := pivot.Format(res)
	if !strings.Contains(out, "[DB_HOST]") {
		t.Errorf("expected section header, got %q", out)
	}
	if !strings.Contains(out, "host1") {
		t.Errorf("expected host1 in output, got %q", out)
	}
}
