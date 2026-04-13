// Package mask provides utilities for redacting sensitive secret values
// before display or logging, based on configurable key patterns.
package mask

import (
	"regexp"
	"strings"
)

// DefaultPatterns are key name patterns that trigger masking by default.
var DefaultPatterns = []string{
	"password", "passwd", "secret", "token", "key", "api_key",
	"private", "credential", "auth", "cert", "dsn",
}

const masked = "***"

// Masker redacts secret values whose keys match configured patterns.
type Masker struct {
	patterns []*regexp.Regexp
}

// New creates a Masker from a list of case-insensitive substring patterns.
func New(patterns []string) (*Masker, error) {
	regs := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(p))
		if err != nil {
			return nil, err
		}
		regs = append(regs, re)
	}
	return &Masker{patterns: regs}, nil
}

// NewDefault returns a Masker using DefaultPatterns.
func NewDefault() *Masker {
	m, _ := New(DefaultPatterns)
	return m
}

// ShouldMask reports whether the given key name matches any pattern.
func (m *Masker) ShouldMask(key string) bool {
	for _, re := range m.patterns {
		if re.MatchString(strings.ToLower(key)) {
			return true
		}
	}
	return false
}

// Apply returns a copy of secrets with sensitive values replaced by "***".
func (m *Masker) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if m.ShouldMask(k) {
			out[k] = masked
		} else {
			out[k] = v
		}
	}
	return out
}

// Value returns the masked representation of a single value given its key.
func (m *Masker) Value(key, value string) string {
	if m.ShouldMask(key) {
		return masked
	}
	return value
}
