package pin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/pin"
)

type mockRW struct {
	data    map[string]map[string]string
	readErr error
	writeErr error
	written map[string]map[string]string
}

func (m *mockRW) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.data[path], nil
}

func (m *mockRW) WriteSecret(_ context.Context, path string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.written == nil {
		m.written = make(map[string]map[string]string)
	}
	m.written[path] = data
	return nil
}

func TestPin_CapturesValues(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/app": {"key": "value"},
	}}
	p := pin.NewPinner(rw, rw)
	entry, err := p.Pin(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Values["key"] != "value" {
		t.Errorf("expected value %q, got %q", "value", entry.Values["key"])
	}
}

func TestPin_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault down")}
	p := pin.NewPinner(rw, rw)
	_, err := p.Pin(context.Background(), "secret/app")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheckDrift_NoDrift(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/app": {"a": "1", "b": "2"},
	}}
	p := pin.NewPinner(rw, rw)
	entry := &pin.PinEntry{Path: "secret/app", Values: map[string]string{"a": "1", "b": "2"}}
	result, err := p.CheckDrift(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Drifted {
		t.Errorf("expected no drift, got diffs: %v", result.Diffs)
	}
}

func TestCheckDrift_Detected(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/app": {"a": "changed", "c": "new"},
	}}
	p := pin.NewPinner(rw, rw)
	entry := &pin.PinEntry{Path: "secret/app", Values: map[string]string{"a": "original", "b": "removed"}}
	result, err := p.CheckDrift(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Drifted {
		t.Fatal("expected drift to be detected")
	}
	if len(result.Diffs) != 3 {
		t.Errorf("expected 3 diffs (changed, removed, added), got %d: %v", len(result.Diffs), result.Diffs)
	}
}

func TestRestore_DryRun(t *testing.T) {
	rw := &mockRW{}
	p := pin.NewPinner(rw, rw)
	entry := &pin.PinEntry{Path: "secret/app", Values: map[string]string{"k": "v"}}
	if err := p.Restore(context.Background(), entry, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rw.written) != 0 {
		t.Error("dry run should not write")
	}
}

func TestRestore_Writes(t *testing.T) {
	rw := &mockRW{}
	p := pin.NewPinner(rw, rw)
	entry := &pin.PinEntry{Path: "secret/app", Values: map[string]string{"k": "v"}}
	if err := p.Restore(context.Background(), entry, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rw.written["secret/app"]["k"] != "v" {
		t.Error("expected restored value")
	}
}

func TestFormatDrift_NoDrift(t *testing.T) {
	out := pin.FormatDrift([]*pin.DriftResult{})
	if out != "no drift detected\n" {
		t.Errorf("unexpected output: %q", out)
	}
}
