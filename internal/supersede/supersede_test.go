package supersede_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/supersede"
)

type mockRW struct {
	data    map[string]map[string]string
	written map[string]map[string]string
	readErr error
}

func (m *mockRW) Read(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	d, ok := m.data[path]
	if !ok {
		return map[string]string{}, nil
	}
	out := make(map[string]string, len(d))
	for k, v := range d {
		out[k] = v
	}
	return out, nil
}

func (m *mockRW) Write(_ context.Context, path string, data map[string]string) error {
	if m.written == nil {
		m.written = make(map[string]map[string]string)
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"src": {"db_pass": "new-secret"},
			"dst": {"db_pass": "old-secret"},
		},
	}
	s := supersede.New(rw, supersede.Options{DryRun: true})
	res, err := s.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if res.Applied() != 1 {
		t.Errorf("expected 1 delta, got %d", res.Applied())
	}
	if rw.written != nil {
		t.Error("expected no writes in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"src": {"api_key": "abc123", "host": "prod.example.com"},
			"dst": {"api_key": "old", "host": "staging.example.com"},
		},
	}
	s := supersede.New(rw, supersede.Options{})
	res, err := s.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Applied() != 2 {
		t.Errorf("expected 2 deltas, got %d", res.Applied())
	}
	if rw.written["dst"]["api_key"] != "abc123" {
		t.Errorf("expected api_key=abc123, got %s", rw.written["dst"]["api_key"])
	}
}

func TestApply_KeyFilter(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"src": {"api_key": "new", "host": "new-host"},
			"dst": {"api_key": "old", "host": "old-host"},
		},
	}
	s := supersede.New(rw, supersede.Options{Keys: []string{"api_key"}})
	res, err := s.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Applied() != 1 {
		t.Errorf("expected 1 delta, got %d", res.Applied())
	}
	if rw.written["dst"]["host"] != "old-host" {
		t.Error("host should not have been superseded")
	}
}

func TestApply_NoChanges(t *testing.T) {
	rw := &mockRW{
		data: map[string]map[string]string{
			"src": {"key": "same"},
			"dst": {"key": "same"},
		},
	}
	s := supersede.New(rw, supersede.Options{})
	res, err := s.Apply(context.Background(), "src", "dst")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Applied() != 0 {
		t.Errorf("expected 0 deltas, got %d", res.Applied())
	}
	if rw.written != nil {
		t.Error("expected no write when nothing changed")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault unavailable")}
	s := supersede.New(rw, supersede.Options{})
	_, err := s.Apply(context.Background(), "src", "dst")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
