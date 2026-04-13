package lint_test

import (
	"testing"

	"github.com/youorg/vaultpatch/internal/lint"
)

func TestRun_NoFindings(t *testing.T) {
	l := lint.New(nil)
	secrets := map[string]string{
		"db_host":     "localhost",
		"db_port":     "5432",
		"api_key":     "supersecretvalue",
	}
	findings := l.Run("secret/myapp", secrets)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestRun_EmptyValue(t *testing.T) {
	l := lint.New(nil)
	secrets := map[string]string{"db_host": ""}
	findings := l.Run("secret/myapp", secrets)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != "no-empty-value" {
		t.Errorf("expected rule no-empty-value, got %s", findings[0].Rule)
	}
	if findings[0].Severity != lint.SeverityError {
		t.Errorf("expected error severity")
	}
}

func TestRun_WhitespaceKey(t *testing.T) {
	l := lint.New(nil)
	secrets := map[string]string{"bad key": "value"}
	findings := l.Run("secret/myapp", secrets)
	found := false
	for _, f := range findings {
		if f.Rule == "no-whitespace-key" {
			found = true
		}
	}
	if !found {
		t.Error("expected no-whitespace-key finding")
	}
}

func TestRun_ShortPassword(t *testing.T) {
	l := lint.New(nil)
	secrets := map[string]string{"db_password": "short"}
	findings := l.Run("secret/myapp", secrets)
	found := false
	for _, f := range findings {
		if f.Rule == "no-plaintext-password" {
			found = true
			if f.Severity != lint.SeverityWarning {
				t.Errorf("expected warning severity")
			}
		}
	}
	if !found {
		t.Error("expected no-plaintext-password finding")
	}
}

func TestRun_KeyNamingConvention(t *testing.T) {
	l := lint.New(nil)
	secrets := map[string]string{"MyKey": "value"}
	findings := l.Run("secret/myapp", secrets)
	found := false
	for _, f := range findings {
		if f.Rule == "key-naming-convention" {
			found = true
		}
	}
	if !found {
		t.Error("expected key-naming-convention finding")
	}
}

func TestHasErrors_False(t *testing.T) {
	findings := []lint.Finding{
		{Rule: "key-naming-convention", Severity: lint.SeverityWarning},
	}
	if lint.HasErrors(findings) {
		t.Error("expected no errors")
	}
}

func TestHasErrors_True(t *testing.T) {
	findings := []lint.Finding{
		{Rule: "no-empty-value", Severity: lint.SeverityError},
	}
	if !lint.HasErrors(findings) {
		t.Error("expected errors")
	}
}

func TestRun_CustomRule(t *testing.T) {
	rules := []lint.Rule{
		{
			Name:     "no-foo",
			Severity: lint.SeverityError,
			Check: func(key, value string) (string, bool) {
				if value == "foo" {
					return "value must not be foo", true
				}
				return "", false
			},
		},
	}
	l := lint.New(rules)
	findings := l.Run("secret/test", map[string]string{"key": "foo"})
	if len(findings) != 1 || findings[0].Rule != "no-foo" {
		t.Errorf("expected no-foo finding, got %+v", findings)
	}
}
