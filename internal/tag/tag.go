// Package tag provides utilities for reading, diffing, and writing
// metadata tags on Vault KV secrets.
package tag

import "fmt"

// Tags is a map of string key/value metadata pairs.
type Tags map[string]string

// Reader is the interface for fetching secret metadata from Vault.
type Reader interface {
	ReadMetadata(path string) (Tags, error)
}

// Writer is the interface for writing secret metadata to Vault.
type Writer interface {
	WriteMetadata(path string, tags Tags) error
}

// Delta represents a single tag change on a secret path.
type Delta struct {
	Path string
	Key  string
	Old  string
	New  string
}

// Compare returns the list of tag deltas between current and desired.
func Compare(current, desired Tags) []Delta {
	var deltas []Delta
	for k, newVal := range desired {
		oldVal, exists := current[k]
		if !exists || oldVal != newVal {
			deltas = append(deltas, Delta{Key: k, Old: oldVal, New: newVal})
		}
	}
	return deltas
}

// ApplyResult holds the outcome of a tag apply operation.
type ApplyResult struct {
	Path    string
	DryRun  bool
	Applied int
	Err     error
}

// Applier applies tag changes to Vault secrets.
type Applier struct {
	writer Writer
	dryRun bool
}

// NewApplier creates a new Applier.
func NewApplier(w Writer, dryRun bool) *Applier {
	return &Applier{writer: w, dryRun: dryRun}
}

// Apply writes desired tags to the given path, merging with current tags.
func (a *Applier) Apply(path string, current, desired Tags) ApplyResult {
	deltas := Compare(current, desired)
	if len(deltas) == 0 {
		return ApplyResult{Path: path, DryRun: a.dryRun}
	}
	if a.dryRun {
		return ApplyResult{Path: path, DryRun: true, Applied: len(deltas)}
	}
	merged := make(Tags, len(current)+len(desired))
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range desired {
		merged[k] = v
	}
	if err := a.writer.WriteMetadata(path, merged); err != nil {
		return ApplyResult{Path: path, Err: fmt.Errorf("write metadata %s: %w", path, err)}
	}
	return ApplyResult{Path: path, Applied: len(deltas)}
}
