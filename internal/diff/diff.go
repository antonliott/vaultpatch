package diff

import (
	"fmt"
	"sort"
	"strings"
)

// SecretMap represents a flat map of secret key-value pairs.
type SecretMap map[string]string

// DiffType indicates the kind of change for a secret key.
type DiffType string

const (
	DiffAdded   DiffType = "added"
	DiffRemoved DiffType = "removed"
	DiffChanged DiffType = "changed"
)

// DiffEntry represents a single difference between two SecretMaps.
type DiffEntry struct {
	Key      string
	Type     DiffType
	OldValue string
	NewValue string
}

// Compare computes the diff between a source and target SecretMap.
// Returns a slice of DiffEntry describing each change.
func Compare(source, target SecretMap) []DiffEntry {
	var entries []DiffEntry

	for k, tv := range target {
		if sv, ok := source[k]; !ok {
			entries = append(entries, DiffEntry{Key: k, Type: DiffAdded, NewValue: tv})
		} else if sv != tv {
			entries = append(entries, DiffEntry{Key: k, Type: DiffChanged, OldValue: sv, NewValue: tv})
		}
	}

	for k, sv := range source {
		if _, ok := target[k]; !ok {
			entries = append(entries, DiffEntry{Key: k, Type: DiffRemoved, OldValue: sv})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return entries
}

// Format renders a human-readable diff output.
func Format(entries []DiffEntry) string {
	if len(entries) == 0 {
		return "No differences found."
	}

	var sb strings.Builder
	for _, e := range entries {
		switch e.Type {
		case DiffAdded:
			sb.WriteString(fmt.Sprintf("+ %s = %q\n", e.Key, e.NewValue))
		case DiffRemoved:
			sb.WriteString(fmt.Sprintf("- %s = %q\n", e.Key, e.OldValue))
		case DiffChanged:
			sb.WriteString(fmt.Sprintf("~ %s: %q -> %q\n", e.Key, e.OldValue, e.NewValue))
		}
	}
	return sb.String()
}
