package version_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/version"
)

func TestGet_ReturnsDefaults(t *testing.T) {
	info := version.Get()

	if info.Version == "" {
		t.Error("expected Version to be non-empty")
	}
	if info.Commit == "" {
		t.Error("expected Commit to be non-empty")
	}
	if info.BuildDate == "" {
		t.Error("expected BuildDate to be non-empty")
	}
}

func TestInfo_String_ContainsVersion(t *testing.T) {
	version.Version = "1.2.3"
	version.Commit = "abc1234"
	version.BuildDate = "2024-01-15"

	info := version.Get()
	s := info.String()

	if !strings.Contains(s, "1.2.3") {
		t.Errorf("expected string to contain version, got: %s", s)
	}
	if !strings.Contains(s, "abc1234") {
		t.Errorf("expected string to contain commit, got: %s", s)
	}
	if !strings.Contains(s, "2024-01-15") {
		t.Errorf("expected string to contain build date, got: %s", s)
	}
	if !strings.HasPrefix(s, "vaultpatch") {
		t.Errorf("expected string to start with 'vaultpatch', got: %s", s)
	}
}

func TestInfo_Short_ReturnsVersion(t *testing.T) {
	version.Version = "2.0.0-beta"

	info := version.Get()
	if info.Short() != "2.0.0-beta" {
		t.Errorf("expected Short() to return '2.0.0-beta', got: %s", info.Short())
	}
}

func TestInfo_String_DevDefaults(t *testing.T) {
	version.Version = "dev"
	version.Commit = "none"
	version.BuildDate = "unknown"

	info := version.Get()
	s := info.String()

	if !strings.Contains(s, "dev") {
		t.Errorf("expected dev build string to contain 'dev', got: %s", s)
	}
}
