// Package pin provides functionality to pin secrets to a specific version
// and detect when they drift from the pinned state.
package pin

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Reader reads secret key-value pairs from a given path.
type Reader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secret key-value pairs to a given path.
type Writer interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// PinEntry records a pinned version of a secret path.
type PinEntry struct {
	Path   string            `json:"path"`
	Values map[string]string `json:"values"`
}

// DriftResult describes drift detected for a single path.
type DriftResult struct {
	Path    string
	Drifted bool
	Diffs   []string
}

// Pinner pins secrets and checks for drift.
type Pinner struct {
	reader Reader
	writer Writer
}

// NewPinner creates a new Pinner.
func NewPinner(r Reader, w Writer) *Pinner {
	return &Pinner{reader: r, writer: w}
}

// Pin captures the current state of path as a PinEntry.
func (p *Pinner) Pin(ctx context.Context, path string) (*PinEntry, error) {
	values, err := p.reader.ReadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("pin: read %q: %w", path, err)
	}
	return &PinEntry{Path: path, Values: values}, nil
}

// CheckDrift compares the current state of path against the pinned entry.
func (p *Pinner) CheckDrift(ctx context.Context, entry *PinEntry) (*DriftResult, error) {
	current, err := p.reader.ReadSecret(ctx, entry.Path)
	if err != nil {
		return nil, fmt.Errorf("pin: drift check read %q: %w", entry.Path, err)
	}

	var diffs []string
	allKeys := unionKeys(entry.Values, current)
	for _, k := range allKeys {
		pinned, inPinned := entry.Values[k]
		live, inLive := current[k]
		switch {
		case inPinned && !inLive:
			diffs = append(diffs, fmt.Sprintf("- %s (removed)", k))
		case !inPinned && inLive:
			diffs = append(diffs, fmt.Sprintf("+ %s (added)", k))
		case pinned != live:
			diffs = append(diffs, fmt.Sprintf("~ %s (changed)", k))
		}
	}

	return &DriftResult{
		Path:    entry.Path,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}, nil
}

// Restore writes the pinned values back to Vault.
func (p *Pinner) Restore(ctx context.Context, entry *PinEntry, dryRun bool) error {
	if dryRun {
		return nil
	}
	if err := p.writer.WriteSecret(ctx, entry.Path, entry.Values); err != nil {
		return fmt.Errorf("pin: restore %q: %w", entry.Path, err)
	}
	return nil
}

func unionKeys(a, b map[string]string) []string {
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

// FormatDrift returns a human-readable drift report.
func FormatDrift(results []*DriftResult) string {
	if len(results) == 0 {
		return "no drift detected\n"
	}
	var sb strings.Builder
	for _, r := range results {
		if !r.Drifted {
			fmt.Fprintf(&sb, "  ok  %s\n", r.Path)
			continue
		}
		fmt.Fprintf(&sb, "DRIFT %s\n", r.Path)
		for _, d := range r.Diffs {
			fmt.Fprintf(&sb, "      %s\n", d)
		}
	}
	return sb.String()
}
