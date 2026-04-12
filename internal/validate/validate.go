// Package validate provides secret value validation against configurable rules.
package validate

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule defines a single validation constraint for a secret key.
type Rule struct {
	Key     string
	Pattern string
	MinLen  int
	Required bool
}

// Violation describes a single failed validation.
type Violation struct {
	Key     string
	Message string
}

func (v Violation) Error() string {
	return fmt.Sprintf("%s: %s", v.Key, v.Message)
}

// Result holds the outcome of a validation run.
type Result struct {
	Violations []Violation
}

// OK returns true when no violations were found.
func (r Result) OK() bool { return len(r.Violations) == 0 }

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	if r.OK() {
		return "all secrets passed validation"
	}
	lines := make([]string, len(r.Violations))
	for i, v := range r.Violations {
		lines[i] = "  " + v.Error()
	}
	return fmt.Sprintf("%d violation(s):\n%s", len(r.Violations), strings.Join(lines, "\n"))
}

// Validator checks a map of secrets against a set of rules.
type Validator struct {
	rules []Rule
}

// NewValidator constructs a Validator with the provided rules.
func NewValidator(rules []Rule) *Validator {
	return &Validator{rules: rules}
}

// Validate runs all rules against secrets and returns a Result.
func (v *Validator) Validate(secrets map[string]string) Result {
	var violations []Violation

	for _, r := range v.rules {
		val, exists := secrets[r.Key]

		if r.Required && !exists {
			violations = append(violations, Violation{Key: r.Key, Message: "required key is missing"})
			continue
		}

		if !exists {
			continue
		}

		if r.MinLen > 0 && len(val) < r.MinLen {
			violations = append(violations, Violation{
				Key:     r.Key,
				Message: fmt.Sprintf("value length %d is below minimum %d", len(val), r.MinLen),
			})
		}

		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				violations = append(violations, Violation{Key: r.Key, Message: fmt.Sprintf("invalid pattern: %v", err)})
				continue
			}
			if !re.MatchString(val) {
				violations = append(violations, Violation{
					Key:     r.Key,
					Message: fmt.Sprintf("value does not match pattern %q", r.Pattern),
				})
			}
		}
	}

	return Result{Violations: violations}
}
