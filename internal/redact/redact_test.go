package redact_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/redact"
)

func TestShouldRedact_MatchesExactKey(t *testing.T) {
	r := redact.New([]string{"password", "token"})
	if !r.ShouldRedact("password") {
		t.Error("expected password to be redacted")
	}
	if !r.ShouldRedact("token") {
		t.Error("expected token to be redacted")
	}
}

func TestShouldRedact_CaseInsensitive(t *testing.T) {
	r := redact.New([]string{"Password"})
	if !r.ShouldRedact("PASSWORD") {
		t.Error("expected case-insensitive match")
	}
}

func TestShouldRedact_NonSensitiveKey(t *testing.T) {
	r := redact.New([]string{"password"})
	if r.ShouldRedact("username") {
		t.Error("username should not be redacted")
	}
}

func TestApply_RedactsSensitiveKeys(t *testing.T) {
	r := redact.New([]string{"secret", "api_key"})
	input := map[string]string{
		"username": "admin",
		"secret":   "s3cr3t",
		"api_key":  "abc123",
	}
	out := r.Apply(input)
	if out["username"] != "admin" {
		t.Errorf("expected admin, got %s", out["username"])
	}
	if out["secret"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %s", out["secret"])
	}
	if out["api_key"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %s", out["api_key"])
	}
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	r := redact.New([]string{"password"})
	input := map[string]string{"password": "hunter2"}
	_ = r.Apply(input)
	if input["password"] != "hunter2" {
		t.Error("Apply must not mutate the input map")
	}
}

func TestValue_RedactedKey(t *testing.T) {
	r := redact.New([]string{"token"})
	if got := r.Value("token", "abc"); got != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %s", got)
	}
}

func TestValue_PlainKey(t *testing.T) {
	r := redact.New([]string{"token"})
	if got := r.Value("host", "localhost"); got != "localhost" {
		t.Errorf("expected localhost, got %s", got)
	}
}

func TestSummary_CountsRedacted(t *testing.T) {
	r := redact.New([]string{"password", "token"})
	secrets := map[string]string{
		"username": "admin",
		"password": "s3cr3t",
		"token":    "tok",
	}
	got := r.Summary(secrets)
	expected := "2 of 3 keys redacted"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
