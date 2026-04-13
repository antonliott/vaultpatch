// Package redact provides functionality to redact secret values
// before writing to logs, output, or external systems.
package redact

import (
	"fmt"
	"strings"
)

const placeholder = "[REDACTED]"

// Redactor replaces secret values with a placeholder string.
type Redactor struct {
	keys map[string]struct{}
}

// New returns a Redactor that redacts values for the given key names.
// Key matching is case-insensitive.
func New(keys []string) *Redactor {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[strings.ToLower(k)] = struct{}{}
	}
	return &Redactor{keys: m}
}

// Apply returns a copy of secrets with sensitive values replaced.
func (r *Redactor) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if r.ShouldRedact(k) {
			out[k] = placeholder
		} else {
			out[k] = v
		}
	}
	return out
}

// ShouldRedact reports whether the given key should be redacted.
func (r *Redactor) ShouldRedact(key string) bool {
	_, ok := r.keys[strings.ToLower(key)]
	return ok
}

// Value returns the value if the key is not redacted, otherwise placeholder.
func (r *Redactor) Value(key, value string) string {
	if r.ShouldRedact(key) {
		return placeholder
	}
	return value
}

// Summary returns a human-readable string listing how many keys were redacted.
func (r *Redactor) Summary(secrets map[string]string) string {
	count := 0
	for k := range secrets {
		if r.ShouldRedact(k) {
			count++
		}
	}
	return fmt.Sprintf("%d of %d keys redacted", count, len(secrets))
}
