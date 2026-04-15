// Package flatten provides utilities for flattening nested Vault secret
// maps into dot-separated key paths, useful for diffing and exporting.
package flatten

import (
	"fmt"
	"sort"
	"strings"
)

// Options controls how flattening is performed.
type Options struct {
	// Separator is placed between path segments. Defaults to ".".
	Separator string
	// Prefix is prepended to every key in the output.
	Prefix string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{Separator: "."}
}

// Flatten converts a nested map[string]any into a flat map[string]string.
// Nested maps are recursed; all other values are converted via fmt.Sprintf.
func Flatten(input map[string]any, opts Options) map[string]string {
	if opts.Separator == "" {
		opts.Separator = "."
	}
	out := make(map[string]string)
	flattenRecurse(input, opts.Prefix, opts.Separator, out)
	return out
}

func flattenRecurse(m map[string]any, prefix, sep string, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + sep + k
		}
		switch child := v.(type) {
		case map[string]any:
			flattenRecurse(child, key, sep, out)
		default:
			out[key] = fmt.Sprintf("%v", v)
		}
	}
}

// Keys returns the sorted keys of a flattened map.
func Keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Unflatten reconstructs a nested map from a flat dot-separated map.
// Only the default "." separator is supported.
func Unflatten(input map[string]string, sep string) map[string]any {
	if sep == "" {
		sep = "."
	}
	out := make(map[string]any)
	for k, v := range input {
		parts := strings.Split(k, sep)
		setNested(out, parts, v)
	}
	return out
}

func setNested(m map[string]any, parts []string, value string) {
	if len(parts) == 1 {
		m[parts[0]] = value
		return
	}
	child, ok := m[parts[0]]
	if !ok {
		child = make(map[string]any)
		m[parts[0]] = child
	}
	if cm, ok := child.(map[string]any); ok {
		setNested(cm, parts[1:], value)
	}
}
