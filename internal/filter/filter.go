package filter

import (
	"regexp"
	"strings"
)

// Rule defines a single filter criterion.
type Rule struct {
	Key   *regexp.Regexp
	Value *regexp.Regexp
}

// Filter holds compiled rules for matching secret entries.
type Filter struct {
	rules []Rule
}

// New creates a Filter from key/value pattern pairs.
// Each pair is "key=value"; omit the value part to match by key only.
func New(patterns []string) (*Filter, error) {
	f := &Filter{}
	for _, p := range patterns {
		parts := strings.SplitN(p, "=", 2)
		kr, err := regexp.Compile(parts[0])
		if err != nil {
			return nil, err
		}
		r := Rule{Key: kr}
		if len(parts) == 2 {
			vr, err := regexp.Compile(parts[1])
			if err != nil {
				return nil, err
			}
			r.Value = vr
		}
		f.rules = append(f.rules, r)
	}
	return f, nil
}

// Match returns true if the key/value pair satisfies at least one rule.
// If no rules are defined, all entries match.
func (f *Filter) Match(key, value string) bool {
	if len(f.rules) == 0 {
		return true
	}
	for _, r := range f.rules {
		if !r.Key.MatchString(key) {
			continue
		}
		if r.Value == nil || r.Value.MatchString(value) {
			return true
		}
	}
	return false
}

// Apply filters a map, returning only entries that match.
func (f *Filter) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if f.Match(k, v) {
			out[k] = v
		}
	}
	return out
}
