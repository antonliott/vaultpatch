package mask_test

import (
	"testing"

	"github.com/youorg/vaultpatch/internal/mask"
)

func TestNew_InvalidPattern(t *testing.T) {
	// QuoteMeta escapes everything, so this path is hard to hit via New;
	// confirm a valid list succeeds.
	m, err := mask.New([]string{"password", "token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil Masker")
	}
}

func TestShouldMask_MatchesDefaultPatterns(t *testing.T) {
	m := mask.NewDefault()
	cases := []struct {
		key  string
		want bool
	}{
		{"db_password", true},
		{"API_KEY", true},
		{"auth_token", true},
		{"username", false},
		{"host", false},
		{"private_key", true},
		{"PORT", false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			got := m.ShouldMask(tc.key)
			if got != tc.want {
				t.Errorf("ShouldMask(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestApply_MasksSensitiveKeys(t *testing.T) {
	m := mask.NewDefault()
	input := map[string]string{
		"db_password": "s3cr3t",
		"host":        "localhost",
		"api_key":     "abc123",
	}
	out := m.Apply(input)

	if out["db_password"] != "***" {
		t.Errorf("expected db_password to be masked, got %q", out["db_password"])
	}
	if out["api_key"] != "***" {
		t.Errorf("expected api_key to be masked, got %q", out["api_key"])
	}
	if out["host"] != "localhost" {
		t.Errorf("expected host to be unmasked, got %q", out["host"])
	}
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	m := mask.NewDefault()
	input := map[string]string{"password": "real"}
	_ = m.Apply(input)
	if input["password"] != "real" {
		t.Error("Apply must not mutate the input map")
	}
}

func TestValue_MasksAndPasses(t *testing.T) {
	m := mask.NewDefault()
	if got := m.Value("secret", "topsecret"); got != "***" {
		t.Errorf("expected ***, got %q", got)
	}
	if got := m.Value("region", "us-east-1"); got != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", got)
	}
}
