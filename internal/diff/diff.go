package diff

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeType represents the kind of difference detected.
type ChangeType string

const (
	Added   ChangeType = "added"
	Removed ChangeType = "removed"
	Changed ChangeType = "changed"
)

// Change describes a single key-level difference between two secret maps.
type Change struct {
	Type     ChangeType
	Key      string
	OldValue string
	NewValue string
}

// Compare returns the list of changes between src and dst secret maps.
func Compare(src, dst map[string]string) []Change {
	var changes []Change

	for k, dv := range dst {
		if sv, ok := src[k]; !ok {
			changes = append(changes, Change{Type: Added, Key: k, NewValue: dv})
		} else if sv != dv {
			changes = append(changes, Change{Type: Changed, Key: k, OldValue: sv, NewValue: dv})
		}
	}

	for k, sv := range src {
		if _, ok := dst[k]; !ok {
			changes = append(changes, Change{Type: Removed, Key: k, OldValue: sv})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Key < changes[j].Key
	})

	return changes
}

// Format returns a human-readable unified-diff-style string for the changes.
func Format(changes []Change) string {
	if len(changes) == 0 {
		return "no changes"
	}
	var sb strings.Builder
	for _, c := range changes {
		switch c.Type {
		case Added:
			fmt.Fprintf(&sb, "+ %s = %s\n", c.Key, c.NewValue)
		case Removed:
			fmt.Fprintf(&sb, "- %s = %s\n", c.Key, c.OldValue)
		case Changed:
			fmt.Fprintf(&sb, "~ %s: %s -> %s\n", c.Key, c.OldValue, c.NewValue)
		}
	}
	return sb.String()
}
