package reorder

import (
	"context"
	"errors"
	"testing"
)

type mockRW struct {
	data    map[string]string
	readErr error
	writeErr error
	written map[string]string
}

func (m *mockRW) ReadSecret(_ context.Context, _ string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	out := make(map[string]string, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *mockRW) WriteSecret(_ context.Context, _ string, data map[string]string) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

func TestApply_AlreadySorted_NoDeltas(t *testing.T) {
	rw := &mockRW{data: map[string]string{"alpha": "1", "beta": "2", "gamma": "3"}}
	r := New(rw, Options{})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Deltas) != 0 {
		t.Fatalf("expected no deltas, got %d", len(res.Deltas))
	}
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]string{"z": "last", "a": "first"}}
	r := New(rw, Options{DryRun: true})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Fatal("expected DryRun=true")
	}
	if rw.written != nil {
		t.Fatal("dry-run should not write")
	}
}

func TestApply_Reverse(t *testing.T) {
	rw := &mockRW{data: map[string]string{"alpha": "1", "beta": "2", "gamma": "3"}}
	r := New(rw, Options{Reverse: true})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	// gamma should now be at index 0
	for _, d := range res.Deltas {
		if d.Key == "gamma" && d.To != 0 {
			t.Fatalf("gamma should be at index 0, got %d", d.To)
		}
	}
}

func TestApply_ExplicitOrder(t *testing.T) {
	rw := &mockRW{data: map[string]string{"c": "3", "a": "1", "b": "2"}}
	r := New(rw, Options{Order: []string{"b", "a", "c"}})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	// With explicit order [b, a, c], key "b" should move to index 0
	for _, d := range res.Deltas {
		if d.Key == "b" && d.To != 0 {
			t.Fatalf("expected b at index 0, got %d", d.To)
		}
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault unavailable")}
	r := New(rw, Options{})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err == nil {
		t.Fatal("expected error")
	}
}

func TestApply_WriteError(t *testing.T) {
	rw := &mockRW{
		data:     map[string]string{"z": "last", "a": "first"},
		writeErr: errors.New("write failed"),
	}
	r := New(rw, Options{})
	res := r.Apply(context.Background(), "secret/test")
	if res.Err == nil {
		t.Fatal("expected write error")
	}
}

func TestFormat_NoDeltas(t *testing.T) {
	out := Format(Result{})
	if out != "no key order changes" {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestFormat_WithDeltas(t *testing.T) {
	res := Result{
		Deltas: []Delta{{Path: "secret/x", Key: "z", From: 0, To: 2}},
	}
	out := Format(res)
	if out == "" {
		t.Fatal("expected non-empty format output")
	}
}
