// Package threshold provides secret count and size threshold checking
// across Vault paths, emitting violations when limits are exceeded.
package threshold

import (
	"context"
	"fmt"
	"sort"
)

// Rule defines a single threshold constraint.
type Rule struct {
	Path     string
	MaxKeys  int // 0 means no limit
	MaxBytes int // 0 means no limit; measured as sum of len(value) for all keys
}

// Violation describes a threshold that was exceeded.
type Violation struct {
	Path    string
	Rule    string
	Actual  int
	Limit   int
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: %s actual=%d limit=%d", v.Path, v.Rule, v.Actual, v.Limit)
}

// SecretReader reads a secret at the given path.
type SecretReader interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Checker evaluates threshold rules against live Vault data.
type Checker struct {
	reader SecretReader
	rules  []Rule
}

// NewChecker creates a Checker with the given reader and rules.
func NewChecker(reader SecretReader, rules []Rule) *Checker {
	return &Checker{reader: reader, rules: rules}
}

// Check evaluates all rules and returns any violations found.
func (c *Checker) Check(ctx context.Context) ([]Violation, error) {
	var violations []Violation

	for _, rule := range c.rules {
		secret, err := c.reader.ReadSecret(ctx, rule.Path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", rule.Path, err)
		}

		if rule.MaxKeys > 0 && len(secret) > rule.MaxKeys {
			violations = append(violations, Violation{
				Path:   rule.Path,
				Rule:   "max_keys",
				Actual: len(secret),
				Limit:  rule.MaxKeys,
			})
		}

		if rule.MaxBytes > 0 {
			total := 0
			for _, v := range secret {
				total += len(v)
			}
			if total > rule.MaxBytes {
				violations = append(violations, Violation{
					Path:   rule.Path,
					Rule:   "max_bytes",
					Actual: total,
					Limit:  rule.MaxBytes,
				})
			}
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Path != violations[j].Path {
			return violations[i].Path < violations[j].Path
		}
		return violations[i].Rule < violations[j].Rule
	})

	return violations, nil
}
