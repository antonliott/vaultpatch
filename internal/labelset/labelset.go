// Package labelset provides functionality for reading, comparing, and
// applying label (metadata) sets to Vault secret paths.
package labelset

import (
	"fmt"
	"sort"
)

// Labels is a map of string key-value metadata pairs.
type Labels map[string]string

// Reader can read labels for a given secret path.
type Reader interface {
	ReadLabels(path string) (Labels, error)
}

// Writer can write labels for a given secret path.
type Writer interface {
	WriteLabels(path string, labels Labels) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Delta describes a single label change.
type Delta struct {
	Key      string
	Old      string
	New      string
	ChangeType string // "added", "removed", "changed"
}

// Compare returns the list of deltas between src and dst label sets.
func Compare(src, dst Labels) []Delta {
	var deltas []Delta
	keys := unionKeys(src, dst)
	for _, k := range keys {
		sv, sinSrc := src[k]
		dv, dinDst := dst[k]
		switch {
		case sinSrc && !dinDst:
			deltas = append(deltas, Delta{Key: k, Old: sv, ChangeType: "removed"})
		case !sinSrc && dinDst:
			deltas = append(deltas, Delta{Key: k, New: dv, ChangeType: "added"})
		case sv != dv:
			deltas = append(deltas, Delta{Key: k, Old: sv, New: dv, ChangeType: "changed"})
		}
	}
	return deltas
}

// Format renders deltas as a human-readable string.
func Format(deltas []Delta) string {
	if len(deltas) == 0 {
		return "no label changes"
	}
	out := ""
	for _, d := range deltas {
		switch d.ChangeType {
		case "added":
			out += fmt.Sprintf("+ %s = %q\n", d.Key, d.New)
		case "removed":
			out += fmt.Sprintf("- %s = %q\n", d.Key, d.Old)
		case "changed":
			out += fmt.Sprintf("~ %s: %q -> %q\n", d.Key, d.Old, d.New)
		}
	}
	return out
}

func unionKeys(a, b Labels) []string {
	seen := make(map[string]struct{})
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
