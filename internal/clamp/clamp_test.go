package clamp_test

import (
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/clamp"
)

type mockRW struct {
	data    map[string]string
	written map[string]string
	readErr error
	writeErr error
}

func (m *mockRW) Read(_ string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	out := make(map[string]string, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockRW) Write(_ string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

func TestApply_NoClamp(t *testing.T) {
	rw := &mockRW{data: map[string]string{"key": "short"}}
	c := clamp.New(rw, clamp.Options{MaxBytes: 10})
	res, err := c.Apply("secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected no results, got %d", len(res))
	}
	if rw.written != nil {
		t.Fatal("expected no write")
	}
}

func TestApply_ClampsLongValue(t *testing.T) {
	rw := &mockRW{data: map[string]string{"token": "abcdefghij_extra"}}
	c := clamp.New(rw, clamp.Options{MaxBytes: 10})
	res, err := c.Apply("secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].OrigLen != 16 || res[0].NewLen != 10 {
		t.Errorf("unexpected lengths: %+v", res[0])
	}
	if rw.written["token"] != "abcdefghij" {
		t.Errorf("unexpected written value: %q", rw.written["token"])
	}
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]string{"secret": "this_is_way_too_long"}}
	c := clamp.New(rw, clamp.Options{MaxBytes: 5, DryRun: true})
	res, err := c.Apply("secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if !res[0].DryRun {
		t.Error("expected DryRun=true")
	}
	if rw.written != nil {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault unavailable")}
	c := clamp.New(rw, clamp.Options{MaxBytes: 32})
	_, err := c.Apply("secret/test")
	if err == nil {
		t.Fatal("expected error")
	}
}
