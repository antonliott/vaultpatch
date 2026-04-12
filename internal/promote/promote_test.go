package promote_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/promote"
)

type mockReader struct {
	data map[string]map[string]interface{}
	err  error
}

func (m *mockReader) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[path], nil
}

type mockWriter struct {
	written map[string]map[string]interface{}
	err     error
}

func (m *mockWriter) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if m.err != nil {
		return m.err
	}
	if m.written == nil {
		m.written = make(map[string]map[string]interface{})
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	src := &mockReader{data: map[string]map[string]interface{}{"secret/a": {"k": "v"}}}
	dst := &mockWriter{}
	p := promote.NewPromoter(src, dst, true)
	results := p.Apply(context.Background(), []string{"secret/a"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Error("expected result to be skipped in dry-run mode")
	}
	if len(dst.written) != 0 {
		t.Error("expected no writes in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	src := &mockReader{data: map[string]map[string]interface{}{"secret/a": {"k": "v"}}}
	dst := &mockWriter{}
	p := promote.NewPromoter(src, dst, false)
	results := p.Apply(context.Background(), []string{"secret/a"})
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if dst.written["secret/a"]["k"] != "v" {
		t.Error("expected secret to be written to destination")
	}
}

func TestApply_ReadError(t *testing.T) {
	src := &mockReader{err: errors.New("vault unavailable")}
	dst := &mockWriter{}
	p := promote.NewPromoter(src, dst, false)
	results := p.Apply(context.Background(), []string{"secret/a"})
	if results[0].Err == nil {
		t.Error("expected read error to be captured")
	}
}

func TestApply_WriteError(t *testing.T) {
	src := &mockReader{data: map[string]map[string]interface{}{"secret/a": {"k": "v"}}}
	dst := &mockWriter{err: errors.New("permission denied")}
	p := promote.NewPromoter(src, dst, false)
	results := p.Apply(context.Background(), []string{"secret/a"})
	if results[0].Err == nil {
		t.Error("expected write error to be captured")
	}
}

func TestSummary_DryRun(t *testing.T) {
	results := []promote.Result{{Path: "a", Skipped: true}, {Path: "b", Skipped: true}}
	got := promote.Summary(results, true)
	expected := "dry-run: 2 path(s) would be promoted, 0 error(s)"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestSummary_WithErrors(t *testing.T) {
	results := []promote.Result{{Path: "a"}, {Path: "b", Err: errors.New("fail")}}
	got := promote.Summary(results, false)
	expected := "promoted 1 path(s), 0 skipped, 1 error(s)"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
