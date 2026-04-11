package patch_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/youorg/vaultpatch/internal/patch"
)

func TestResult_HasErrors_False(t *testing.T) {
	r := &patch.Result{Path: "secret/app", Applied: 3}
	if r.HasErrors() {
		t.Fatal("expected no errors")
	}
}

func TestResult_HasErrors_True(t *testing.T) {
	r := &patch.Result{Path: "secret/app"}
	r.AddError(errors.New("something failed"))
	if !r.HasErrors() {
		t.Fatal("expected errors")
	}
}

func TestResult_Summary_DryRun(t *testing.T) {
	r := &patch.Result{
		Path:    "secret/app",
		Applied: 2,
		Skipped: 1,
		DryRun:  true,
	}
	s := r.Summary()
	if !strings.Contains(s, "dry-run") {
		t.Errorf("expected dry-run in summary, got: %s", s)
	}
	if !strings.Contains(s, "applied=2") {
		t.Errorf("expected applied=2 in summary, got: %s", s)
	}
}

func TestResult_Summary_WithErrors(t *testing.T) {
	r := &patch.Result{Path: "secret/app", Applied: 0}
	r.AddError(errors.New("write failed"))
	s := r.Summary()
	if !strings.Contains(s, "errors=1") {
		t.Errorf("expected errors=1 in summary, got: %s", s)
	}
	if !strings.Contains(s, "write failed") {
		t.Errorf("expected error message in summary, got: %s", s)
	}
}

func TestResult_Summary_Applied(t *testing.T) {
	r := &patch.Result{Path: "secret/db", Applied: 5, Skipped: 0, DryRun: false}
	s := r.Summary()
	if !strings.Contains(s, "applied") {
		t.Errorf("expected 'applied' mode in summary, got: %s", s)
	}
}
