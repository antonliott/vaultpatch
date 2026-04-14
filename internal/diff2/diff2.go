// Package diff2 provides two-way diffing of Vault secret paths between
// two namespaces, producing a structured delta suitable for review or apply.
package diff2

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeType classifies a single key-level change.
type ChangeType string

const (
	Added   ChangeType = "added"
	Removed ChangeType = "removed"
	Changed ChangeType = "changed"
)

// Delta represents a single changed key within a secret path.
type Delta struct {
	Path      string
	Key       string
	Type      ChangeType
	OldValue  string
	NewValue  string
}

// Result holds all deltas produced by a comparison.
type Result struct {
	Deltas []Delta
}

// HasChanges reports whether any deltas were found.
func (r Result) HasChanges() bool {
	return len(r.Deltas) > 0
}

// Compare diffs two maps of path→(key→value) and returns a Result.
func Compare(src, dst map[string]map[string]string) Result {
	var deltas []Delta

	allPaths := unionKeys(src, dst)
	for _, path := range allPaths {
		srcSecrets := src[path]
		dstSecrets := dst[path]
		allKeys := unionKeys(srcSecrets, dstSecrets)
		for _, key := range allKeys {
			sv, srcOk := srcSecrets[key]
			dv, dstOk := dstSecrets[key]
			switch {
			case srcOk && !dstOk:
				deltas = append(deltas, Delta{Path: path, Key: key, Type: Removed, OldValue: sv})
			case !srcOk && dstOk:
				deltas = append(deltas, Delta{Path: path, Key: key, Type: Added, NewValue: dv})
			case sv != dv:
				deltas = append(deltas, Delta{Path: path, Key: key, Type: Changed, OldValue: sv, NewValue: dv})
			}
		}
	}
	return Result{Deltas: deltas}
}

// Format renders a Result as a human-readable diff string.
func Format(r Result) string {
	if !r.HasChanges() {
		return "no differences found"
	}
	var sb strings.Builder
	for _, d := range r.Deltas {
		switch d.Type {
		case Added:
			fmt.Fprintf(&sb, "+ %s#%s = %q\n", d.Path, d.Key, d.NewValue)
		case Removed:
			fmt.Fprintf(&sb, "- %s#%s (was %q)\n", d.Path, d.Key, d.OldValue)
		case Changed:
			fmt.Fprintf(&sb, "~ %s#%s: %q → %q\n", d.Path, d.Key, d.OldValue, d.NewValue)
		}
	}
	return sb.String()
}

func unionKeys[V any](a, b map[string]V) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
