package health_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/health"
)

func TestStatus_String_Unsealed(t *testing.T) {
	s := health.Status{
		Address:     "http://127.0.0.1:8200",
		Namespace:   "ns1",
		Initialized: true,
		Sealed:      false,
		Version:     "1.15.0",
		CheckedAt:   time.Now(),
	}
	got := s.String()
	if got == "" {
		t.Fatal("expected non-empty status string")
	}
	for _, want := range []string{"1.15.0", "http://127.0.0.1:8200", "ns1", "initialized", "unsealed"} {
		if !contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}

func TestStatus_String_Sealed(t *testing.T) {
	s := health.Status{
		Address:     "http://vault:8200",
		Namespace:   "",
		Initialized: true,
		Sealed:      true,
		Version:     "1.14.2",
		CheckedAt:   time.Now(),
	}
	got := s.String()
	if !contains(got, "sealed") {
		t.Errorf("String() = %q, want it to contain \"sealed\"", got)
	}
}

func TestStatus_String_Uninitialized(t *testing.T) {
	s := health.Status{
		Address:     "http://vault:8200",
		Namespace:   "dev",
		Initialized: false,
		Sealed:      true,
		Version:     "1.13.0",
		CheckedAt:   time.Now(),
	}
	got := s.String()
	if !contains(got, "uninitialized") {
		t.Errorf("String() = %q, want it to contain \"uninitialized\"", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
