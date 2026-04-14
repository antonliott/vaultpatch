package tokenize_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/tokenize"
)

func TestToken_IsStable(t *testing.T) {
	tz := tokenize.New("VAULT")
	tok1, err := tz.Token("s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok2, err := tz.Token("s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok1 != tok2 {
		t.Errorf("expected stable token, got %q then %q", tok1, tok2)
	}
}

func TestToken_UsesPrefix(t *testing.T) {
	tz := tokenize.New("MYAPP")
	tok, err := tz.Token("value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(tok, "<MYAPP_") {
		t.Errorf("expected prefix <MYAPP_, got %q", tok)
	}
}

func TestToken_DefaultPrefix(t *testing.T) {
	tz := tokenize.New("")
	tok, err := tz.Token("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(tok, "<TOKEN_") {
		t.Errorf("expected default prefix <TOKEN_, got %q", tok)
	}
}

func TestApply_ReplacesValues(t *testing.T) {
	tz := tokenize.New("V")
	secrets := map[string]string{
		"db_pass": "hunter2",
		"api_key": "abc123",
	}
	out, err := tz.Apply(secrets)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	for k, v := range out {
		if v == secrets[k] {
			t.Errorf("key %q: value was not tokenized", k)
		}
		if !strings.HasPrefix(v, "<V_") {
			t.Errorf("key %q: unexpected token format %q", k, v)
		}
	}
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	tz := tokenize.New("V")
	orig := map[string]string{"k": "original"}
	_, err := tz.Apply(orig)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if orig["k"] != "original" {
		t.Error("Apply mutated the input map")
	}
}

func TestDetokenize_RestoresValues(t *testing.T) {
	tz := tokenize.New("V")
	tok, _ := tz.Token("s3cr3t")
	text := "connect using " + tok + " now"
	got := tz.Detokenize(text)
	if !strings.Contains(got, "s3cr3t") {
		t.Errorf("expected restored value in %q", got)
	}
	if strings.Contains(got, tok) {
		t.Errorf("token should be replaced in %q", got)
	}
}

func TestDetokenize_UnknownTokenUnchanged(t *testing.T) {
	tz := tokenize.New("V")
	text := "no known tokens here <V_deadbeef>"
	got := tz.Detokenize(text)
	if got != text {
		t.Errorf("expected unchanged text, got %q", got)
	}
}

func TestLen_CountsDistinctValues(t *testing.T) {
	tz := tokenize.New("V")
	_, _ = tz.Token("a")
	_, _ = tz.Token("b")
	_, _ = tz.Token("a") // duplicate
	if tz.Len() != 2 {
		t.Errorf("expected 2, got %d", tz.Len())
	}
}
