// Package drift detects configuration drift between two Vault secret paths
// by comparing their current live values against a previously captured baseline.
package drift

import (
	"context"
	"fmt"
	"sort"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Delta represents a single key-level drift finding.
type Delta struct {
	Key      string
	Baseline string
	Current  string
	Status   string // "changed", "added", "removed"
}

// Report holds all detected deltas for a path.
type Report struct {
	Path   string
	Deltas []Delta
}

// HasDrift returns true when at least one delta was detected.
func (r Report) HasDrift() bool { return len(r.Deltas) > 0 }

// Summary returns a short human-readable summary.
func (r Report) Summary() string {
	if !r.HasDrift() {
		return fmt.Sprintf("%s: no drift detected", r.Path)
	}
	return fmt.Sprintf("%s: %d drift delta(s) detected", r.Path, len(r.Deltas))
}

// Detector compares a baseline snapshot against live Vault secrets.
type Detector struct {
	reader Reader
}

// NewDetector creates a Detector backed by the given Reader.
func NewDetector(r Reader) *Detector {
	return &Detector{reader: r}
}

// Detect reads the current secrets at path and compares them to baseline.
func (d *Detector) Detect(ctx context.Context, path string, baseline map[string]string) (Report, error) {
	current, err := d.reader.Read(ctx, path)
	if err != nil {
		return Report{}, fmt.Errorf("drift: read %s: %w", path, err)
	}

	deltas := compare(baseline, current)
	sort.Slice(deltas, func(i, j int) bool { return deltas[i].Key < deltas[j].Key })

	return Report{Path: path, Deltas: deltas}, nil
}

func compare(baseline, current map[string]string) []Delta {
	seen := make(map[string]bool)
	var deltas []Delta

	for k, bv := range baseline {
		seen[k] = true
		cv, ok := current[k]
		if !ok {
			deltas = append(deltas, Delta{Key: k, Baseline: bv, Current: "", Status: "removed"})
		} else if bv != cv {
			deltas = append(deltas, Delta{Key: k, Baseline: bv, Current: cv, Status: "changed"})
		}
	}

	for k, cv := range current {
		if !seen[k] {
			deltas = append(deltas, Delta{Key: k, Baseline: "", Current: cv, Status: "added"})
		}
	}

	return deltas
}
