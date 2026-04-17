package filter_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/filter"
)

func TestNew_InvalidKeyPattern(t *testing.T) {
	_, err := filter.New([]string{"[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid key pattern")
	}
}

func TestNew_InvalidValuePattern(t *testing.T) {
	_, err := filter.New([]string{"key=[bad"})
	if err == nil {
		t.Fatal("expected error for invalid value pattern")
	}
}

func TestMatch_NoRules_AlwaysTrue(t *testing.T) {
	f, _ := filter.New(nil)
	if !f.Match("any", "value") {
		t.Fatal("expected match with no rules")
	}
}

func TestMatch_KeyOnly(t *testing.T) {
	f, _ := filter.New([]string{"^db_"})
	if !f.Match("db_password", "secret") {
		t.Error("expected match")
	}
	if f.Match("api_key", "secret") {
		t.Error("unexpected match")
	}
}

func TestMatch_KeyAndValue(t *testing.T) {
	f, _ := filter.New([]string{"password=^change"})
	if !f.Match("db_password", "changeme") {
		t.Error("expected match")
	}
	if f.Match("db_password", "hunter2") {
		t.Error("unexpected match on value")
	}
}

func TestApply_FiltersMap(t *testing.T) {
	f, _ := filter.New([]string{"^token"})
	input := map[string]string{
		"token_a": "aaa",
		"token_b": "bbb",
		"password": "secret",
	}
	out := f.Apply(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if _, ok := out["password"]; ok {
		t.Error("password should have been filtered out")
	}
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	f, _ := filter.New([]string{"keep"})
	input := map[string]string{"keep_me": "yes", "drop_me": "no"}
	_ = f.Apply(input)
	if len(input) != 2 {
		t.Error("input map was mutated")
	}
}
