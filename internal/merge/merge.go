// Package merge provides functionality for merging secrets from multiple
// Vault paths into a single destination path.
package merge

import (
	"context"
	"fmt"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	ReadSecrets(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secrets to a Vault path.
type Writer interface {
	WriteSecrets(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter can both read and write secrets.
type ReadWriter interface {
	Reader
	Writer
}

// Options controls merge behaviour.
type Options struct {
	// DryRun prints what would change without writing.
	DryRun bool
	// Overwrite determines whether existing keys at the destination are
	// overwritten when a source key conflicts.
	Overwrite bool
}

// Result summarises a completed merge operation.
type Result struct {
	Merged  int
	Skipped int
	Err     error
}

// Merger merges secrets from one or more source paths into a destination.
type Merger struct {
	rw  ReadWriter
	opts Options
}

// NewMerger creates a Merger with the given ReadWriter and Options.
func NewMerger(rw ReadWriter, opts Options) *Merger {
	return &Merger{rw: rw, opts: opts}
}

// Apply reads secrets from each source path and merges them into dst.
// Sources are applied in order; later sources win on conflict only when
// Overwrite is true.
func (m *Merger) Apply(ctx context.Context, dst string, sources []string) (Result, error) {
	merged := make(map[string]string)

	// Seed with existing destination secrets so non-conflicting keys survive.
	existing, err := m.rw.ReadSecrets(ctx, dst)
	if err != nil {
		return Result{Err: err}, fmt.Errorf("merge: read destination %q: %w", dst, err)
	}
	for k, v := range existing {
		merged[k] = v
	}

	var mergedCount, skippedCount int
	for _, src := range sources {
		secrets, err := m.rw.ReadSecrets(ctx, src)
		if err != nil {
			return Result{Err: err}, fmt.Errorf("merge: read source %q: %w", src, err)
		}
		for k, v := range secrets {
			if _, exists := merged[k]; exists && !m.opts.Overwrite {
				skippedCount++
				continue
			}
			merged[k] = v
			mergedCount++
		}
	}

	if !m.opts.DryRun {
		if err := m.rw.WriteSecrets(ctx, dst, merged); err != nil {
			return Result{Err: err}, fmt.Errorf("merge: write destination %q: %w", dst, err)
		}
	}

	return Result{Merged: mergedCount, Skipped: skippedCount}, nil
}
