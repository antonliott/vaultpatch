// Package normalize provides key normalization for Vault secret maps.
// It standardizes key casing, trims whitespace, and optionally replaces
// separator characters to produce consistent key formats across namespaces.
package normalize

import (
	"fmt"
	"regexp"
	"strings"
)

// Style defines the target key normalization style.
type Style string

const (
	StyleUpper  Style = "upper"
	StyleLower  Style = "lower"
	StyleSnake  Style = "snake"
	StyleKebab  Style = "kebab"
)

var validStyles = map[Style]bool{
	StyleUpper: true,
	StyleLower: true,
	StyleSnake: true,
	StyleKebab: true,
}

var sepRe = regexp.MustCompile(`[\s\-_]+`)

// Normalizer transforms secret map keys according to a configured style.
type Normalizer struct {
	style Style
}

// New returns a Normalizer for the given style, or an error if the style
// is not recognised.
func New(style Style) (*Normalizer, error) {
	if !validStyles[style] {
		return nil, fmt.Errorf("normalize: unknown style %q; valid values: upper, lower, snake, kebab")
	}
	return &Normalizer{style: style}, nil
}

// Apply returns a new map with all keys normalised. Values are unchanged.
// The original map is never mutated.
func (n *Normalizer) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[n.normalizeKey(k)] = v
	}
	return out
}

func (n *Normalizer) normalizeKey(k string) string {
	k = strings.TrimSpace(k)
	switch n.style {
	case StyleUpper:
		return strings.ToUpper(k)
	case StyleLower:
		return strings.ToLower(k)
	case StyleSnake:
		return strings.ToLower(sepRe.ReplaceAllString(k, "_"))
	case StyleKebab:
		return strings.ToLower(sepRe.ReplaceAllString(k, "-"))
	}
	return k
}
