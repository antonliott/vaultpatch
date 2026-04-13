// Package lint provides rules-based validation of Vault secret keys and values.
package lint

import (
	"fmt"
	"regexp"
	"strings"
)

// Severity indicates the importance of a lint finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Finding represents a single lint result for a key/value pair.
type Finding struct {
	Path     string
	Key      string
	Rule     string
	Message  string
	Severity Severity
}

// Rule defines a single lint check.
type Rule struct {
	Name     string
	Severity Severity
	Check    func(key, value string) (string, bool)
}

// Linter runs a set of rules against a map of secrets.
type Linter struct {
	rules []Rule
}

// New returns a Linter loaded with the provided rules.
// If rules is empty, DefaultRules are used.
func New(rules []Rule) *Linter {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	return &Linter{rules: rules}
}

// DefaultRules returns the built-in lint rules.
func DefaultRules() []Rule {
	return []Rule{
		{
			Name:     "no-empty-value",
			Severity: SeverityError,
			Check: func(key, value string) (string, bool) {
				if strings.TrimSpace(value) == "" {
					return "value must not be empty", true
				}
				return "", false
			},
		},
		{
			Name:     "no-whitespace-key",
			Severity: SeverityError,
			Check: func(key, value string) (string, bool) {
				if strings.ContainsAny(key, " \t") {
					return "key must not contain whitespace", true
				}
				return "", false
			},
		},
		{
			Name:     "no-plaintext-password",
			Severity: SeverityWarning,
			Check: func(key, value string) (string, bool) {
				lower := strings.ToLower(key)
				if (strings.Contains(lower, "password") || strings.Contains(lower, "passwd")) &&
					len(value) < 12 {
					return "password value appears short (< 12 chars)", true
				}
				return "", false
			},
		},
		{
			Name:     "key-naming-convention",
			Severity: SeverityWarning,
			Check: func(key, value string) (string, bool) {
				ok, _ := regexp.MatchString(`^[a-z][a-z0-9_]*$`, key)
				if !ok {
					return fmt.Sprintf("key %q should match ^[a-z][a-z0-9_]*$", key), true
				}
				return "", false
			},
		},
	}
}

// Run applies all rules to the provided secrets map and returns findings.
func (l *Linter) Run(path string, secrets map[string]string) []Finding {
	var findings []Finding
	for key, value := range secrets {
		for _, rule := range l.rules {
			if msg, triggered := rule.Check(key, value); triggered {
				findings = append(findings, Finding{
					Path:     path,
					Key:      key,
					Rule:     rule.Name,
					Message:  msg,
					Severity: rule.Severity,
				})
			}
		}
	}
	return findings
}

// HasErrors returns true if any finding has severity error.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}
