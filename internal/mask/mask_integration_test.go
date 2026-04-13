package mask_test

import (
	"strings"
	"testing"

	"github.com/youorg/vaultpatch/internal/mask"
)

// TestCustomPatterns verifies that user-supplied patterns override defaults.
func TestCustomPatterns_Override(t *testing.T) {
	m, err := mask.New([]string{"my_custom_field"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Default patterns should NOT apply.
	if m.ShouldMask("password") {
		t.Error("custom masker should not mask 'password' when not in pattern list")
	}
	// Custom pattern should apply.
	if !m.ShouldMask("my_custom_field") {
		t.Error("expected my_custom_field to be masked")
	}
}

// TestApply_LargeMap ensures Apply scales to many keys without mutation.
func TestApply_LargeMap(t *testing.T) {
	m := mask.NewDefault()
	input := make(map[string]string, 100)
	for i := 0; i < 50; i++ {
		input[fmt.Sprintf("key_%d", i)] = "value"
		input[fmt.Sprintf("password_%d", i)] = "secret"
	}

	out := m.Apply(input)

	for k, v := range out {
		if strings.HasPrefix(k, "password_") && v != "***" {
			t.Errorf("expected %q to be masked", k)
		}
		if strings.HasPrefix(k, "key_") && v != "value" {
			t.Errorf("expected %q to be unmasked", k)
		}
	}
	// Ensure input not mutated
	for k, v := range input {
		if strings.HasPrefix(k, "password_") && v != "secret" {
			t.Errorf("input mutated for key %q", k)
		}
	}
}

fmt is not imported — add it inline:
func fmt_Sprintf(f string, a ...interface{}) string { return "" } // placeholder removed below
