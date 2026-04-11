package vault

import (
	"testing"
)

func TestKvMount(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"secret/myapp/config", "secret"},
		{"kv/service/db", "kv"},
		{"secret", "secret"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := kvMount(tt.path)
			if got != tt.expected {
				t.Errorf("kvMount(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestKvKey(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"secret/myapp/config", "myapp/config"},
		{"kv/service/db", "service/db"},
		{"secret", "secret"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := kvKey(tt.path)
			if got != tt.expected {
				t.Errorf("kvKey(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestKvV2DataPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"secret/myapp/config", "secret/data/myapp/config"},
		{"kv/svc", "kv/data/svc"},
		{"secret", "secret"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := kvV2DataPath(tt.path)
			if got != tt.expected {
				t.Errorf("kvV2DataPath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
