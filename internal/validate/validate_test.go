package validate_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/validate"
)

func TestValidate_AllPass(t *testing.T) {
	v := validate.NewValidator([]validate.Rule{
		{Key: "DB_PASSWORD", Required: true, MinLen: 8},
		{Key: "API_KEY", Pattern: `^[A-Za-z0-9]{16,}$`},
	})
	secrets := map[string]string{
		"DB_PASSWORD": "supersecret",
		"API_KEY":     "AbCdEfGhIjKlMnOp",
	}
	res := v.Validate(secrets)
	if !res.OK() {
		t.Fatalf("expected no violations, got: %s", res.Summary())
	}
}

func TestValidate_RequiredMissing(t *testing.T) {
	v := validate.NewValidator([]validate.Rule{
		{Key: "TOKEN", Required: true},
	})
	res := v.Validate(map[string]string{})
	if res.OK() {
		t.Fatal("expected violation for missing required key")
	}
	if res.Violations[0].Key != "TOKEN" {
		t.Errorf("unexpected violation key: %s", res.Violations[0].Key)
	}
}

func TestValidate_MinLenViolation(t *testing.T) {
	v := validate.NewValidator([]validate.Rule{
		{Key: "SECRET", MinLen: 12},
	})
	res := v.Validate(map[string]string{"SECRET": "short"})
	if res.OK() {
		t.Fatal("expected min-length violation")
	}
	if !strings.Contains(res.Violations[0].Message, "below minimum") {
		t.Errorf("unexpected message: %s", res.Violations[0].Message)
	}
}

func TestValidate_PatternViolation(t *testing.T) {
	v := validate.NewValidator([]validate.Rule{
		{Key: "PORT", Pattern: `^\d+$`},
	})
	res := v.Validate(map[string]string{"PORT": "not-a-number"})
	if res.OK() {
		t.Fatal("expected pattern violation")
	}
	if !strings.Contains(res.Violations[0].Message, "does not match pattern") {
		t.Errorf("unexpected message: %s", res.Violations[0].Message)
	}
}

func TestValidate_OptionalKeyAbsent(t *testing.T) {
	v := validate.NewValidator([]validate.Rule{
		{Key: "OPTIONAL", MinLen: 5},
	})
	res := v.Validate(map[string]string{})
	if !res.OK() {
		t.Fatalf("optional absent key should not produce violation: %s", res.Summary())
	}
}

func TestResult_Summary_WithViolations(t *testing.T) {
	res := validate.Result{
		Violations: []validate.Violation{
			{Key: "A", Message: "required key is missing"},
			{Key: "B", Message: "too short"},
		},
	}
	summary := res.Summary()
	if !strings.Contains(summary, "2 violation(s)") {
		t.Errorf("expected violation count in summary, got: %s", summary)
	}
}

func TestResult_Summary_OK(t *testing.T) {
	res := validate.Result{}
	if res.Summary() != "all secrets passed validation" {
		t.Errorf("unexpected summary: %s", res.Summary())
	}
}
