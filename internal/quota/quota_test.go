package quota_test

import (
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/quota"
)

type mockReader struct {
	data map[string][]string
	err  error
}

func (m *mockReader) List(path string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[path], nil
}

func TestCheck_NoViolations(t *testing.T) {
	r := &mockReader{data: map[string][]string{
		"secret/app": {"key1", "key2"},
	}}
	c := quota.NewChecker(r, []quota.Rule{{Path: "secret/app", Limit: 5}})
	v, err := c.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %d", len(v))
	}
}

func TestCheck_Violation(t *testing.T) {
	r := &mockReader{data: map[string][]string{
		"secret/app": {"k1", "k2", "k3", "k4"},
	}}
	c := quota.NewChecker(r, []quota.Rule{{Path: "secret/app", Limit: 2}})
	v, err := c.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Actual != 4 || v[0].Limit != 2 {
		t.Errorf("unexpected violation: %+v", v[0])
	}
}

func TestCheck_ListError(t *testing.T) {
	r := &mockReader{err: errors.New("connection refused")}
	c := quota.NewChecker(r, []quota.Rule{{Path: "secret/app", Limit: 10}})
	_, err := c.Check()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheck_MultipleRules_SortedOutput(t *testing.T) {
	r := &mockReader{data: map[string][]string{
		"secret/z": {"a", "b", "c"},
		"secret/a": {"x", "y", "z"},
	}}
	rules := []quota.Rule{
		{Path: "secret/z", Limit: 1},
		{Path: "secret/a", Limit: 1},
	}
	c := quota.NewChecker(r, rules)
	v, err := c.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
	if v[0].Path != "secret/a" || v[1].Path != "secret/z" {
		t.Errorf("violations not sorted: %v, %v", v[0].Path, v[1].Path)
	}
}
