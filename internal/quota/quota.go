// Package quota provides tools for inspecting and comparing secret counts
// against configurable per-path quotas in Vault namespaces.
package quota

import (
	"fmt"
	"sort"
)

// Reader lists secret keys under a given path.
type Reader interface {
	List(path string) ([]string, error)
}

// Rule defines a quota limit for a path prefix.
type Rule struct {
	Path  string
	Limit int
}

// Violation records a path that has exceeded its quota.
type Violation struct {
	Path    string
	Limit   int
	Actual  int
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: limit %d, actual %d", v.Path, v.Limit, v.Actual)
}

// Checker evaluates secret counts against quota rules.
type Checker struct {
	reader Reader
	rules  []Rule
}

// NewChecker creates a Checker with the given reader and rules.
func NewChecker(r Reader, rules []Rule) *Checker {
	return &Checker{reader: r, rules: rules}
}

// Check evaluates all rules and returns any violations found.
func (c *Checker) Check() ([]Violation, error) {
	var violations []Violation

	for _, rule := range c.rules {
		keys, err := c.reader.List(rule.Path)
		if err != nil {
			return nil, fmt.Errorf("quota: list %q: %w", rule.Path, err)
		}
		if len(keys) > rule.Limit {
			violations = append(violations, Violation{
				Path:   rule.Path,
				Limit:  rule.Limit,
				Actual: len(keys),
			})
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].Path < violations[j].Path
	})
	return violations, nil
}
