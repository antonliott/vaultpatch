package threshold_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/threshold"
)

type mockReader struct {
	data    map[string]map[string]string
	readErr error
}

func (m *mockReader) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (m *mockReader) ReadSecret(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.data[path], nil
}

func TestCheck_NoViolations(t *testing.T) {
	reader := &mockReader{data: map[string]map[string]string{
		"secret/app": {"key1": "val1", "key2": "val2"},
	}}
	checker := threshold.NewChecker(reader, []threshold.Rule{
		{Path: "secret/app", MaxKeys: 5, MaxBytes: 1000},
	})
	violations, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestCheck_MaxKeysViolation(t *testing.T) {
	reader := &mockReader{data: map[string]map[string]string{
		"secret/app": {"a": "1", "b": "2", "c": "3"},
	}}
	checker := threshold.NewChecker(reader, []threshold.Rule{
		{Path: "secret/app", MaxKeys: 2},
	})
	violations, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Rule != "max_keys" {
		t.Errorf("expected max_keys rule, got %q", violations[0].Rule)
	}
	if violations[0].Actual != 3 || violations[0].Limit != 2 {
		t.Errorf("unexpected actual/limit: %d/%d", violations[0].Actual, violations[0].Limit)
	}
}

func TestCheck_MaxBytesViolation(t *testing.T) {
	reader := &mockReader{data: map[string]map[string]string{
		"secret/big": {"k": "this-is-a-long-value-exceeding-limit"},
	}}
	checker := threshold.NewChecker(reader, []threshold.Rule{
		{Path: "secret/big", MaxBytes: 10},
	})
	violations, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(violations) != 1 || violations[0].Rule != "max_bytes" {
		t.Fatalf("expected max_bytes violation, got %+v", violations)
	}
}

func TestCheck_ReadError(t *testing.T) {
	reader := &mockReader{readErr: errors.New("vault unavailable")}
	checker := threshold.NewChecker(reader, []threshold.Rule{
		{Path: "secret/app", MaxKeys: 1},
	})
	_, err := checker.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestViolation_String(t *testing.T) {
	v := threshold.Violation{Path: "secret/app", Rule: "max_keys", Actual: 5, Limit: 3}
	s := v.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
