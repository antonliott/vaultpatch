// Package policy provides utilities for comparing and formatting
// HashiCorp Vault ACL policy differences between namespaces.
package policy

import (
	"fmt"
	"sort"
	"strings"
)

// DiffType indicates the kind of change detected for a policy.
type DiffType string

const (
	DiffAdded   DiffType = "added"
	DiffRemoved DiffType = "removed"
	DiffChanged DiffType = "changed"
)

// Delta represents a single policy difference.
type Delta struct {
	Name    string
	Type    DiffType
	OldRules string
	NewRules string
}

// Compare returns the list of policy deltas between two maps of
// policy name -> HCL rules strings.
func Compare(source, target map[string]string) []Delta {
	var deltas []Delta

	// Check for added or changed policies.
	keys := make([]string, 0, len(source))
	for k := range source {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		srcRules := source[name]
		tgtRules, exists := target[name]
		if !exists {
			deltas = append(deltas, Delta{Name: name, Type: DiffAdded, NewRules: srcRules})
		} else if strings.TrimSpace(srcRules) != strings.TrimSpace(tgtRules) {
			deltas = append(deltas, Delta{Name: name, Type: DiffChanged, OldRules: tgtRules, NewRules: srcRules})
		}
	}

	// Check for removed policies.
	tgtKeys := make([]string, 0, len(target))
	for k := range target {
		tgtKeys = append(tgtKeys, k)
	}
	sort.Strings(tgtKeys)

	for _, name := range tgtKeys {
		if _, exists := source[name]; !exists {
			deltas = append(deltas, Delta{Name: name, Type: DiffRemoved, OldRules: target[name]})
		}
	}

	return deltas
}

// Format renders a human-readable summary of policy deltas.
func Format(deltas []Delta) string {
	if len(deltas) == 0 {
		return "no policy differences found"
	}
	var sb strings.Builder
	for _, d := range deltas {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", d.Type, d.Name))
	}
	return strings.TrimRight(sb.String(), "\n")
}
