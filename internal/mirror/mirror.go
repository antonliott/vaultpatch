// Package mirror copies all secrets from one Vault path prefix to another,
// optionally across namespaces.
package mirror

import (
	"context"
	"fmt"
)

// Reader reads secret keys and values from a Vault path.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secret values to a Vault path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]string) error
}

// Result holds the outcome of a mirror operation.
type Result struct {
	Mirrored []string
	Skipped  []string
	Errors   []error
	DryRun   bool
}

// HasErrors returns true if any errors occurred.
func (r *Result) HasErrors() bool { return len(r.Errors) > 0 }

// Summary returns a human-readable summary of the result.
func (r *Result) Summary() string {
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would mirror %d path(s), skip %d", len(r.Mirrored), len(r.Skipped))
	}
	return fmt.Sprintf("mirrored %d path(s), skipped %d, errors %d", len(r.Mirrored), len(r.Skipped), len(r.Errors))
}

// Mirrorer mirrors secrets from a source prefix to a destination prefix.
type Mirrorer struct {
	reader Reader
	writer Writer
	overwrite bool
	dryRun    bool
}

// NewMirrorer creates a new Mirrorer.
func NewMirrorer(r Reader, w Writer, overwrite, dryRun bool) *Mirrorer {
	return &Mirrorer{reader: r, writer: w, overwrite: overwrite, dryRun: dryRun}
}

// Apply mirrors all secrets from src prefix to dst prefix.
func (m *Mirrorer) Apply(ctx context.Context, src, dst string) (*Result, error) {
	keys, err := m.reader.List(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("mirror: list %q: %w", src, err)
	}

	res := &Result{DryRun: m.dryRun}

	for _, key := range keys {
		srcPath := src + "/" + key
		dstPath := dst + "/" + key

		data, err := m.reader.Read(ctx, srcPath)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("read %q: %w", srcPath, err))
			continue
		}

		if !m.overwrite {
			existing, err := m.reader.Read(ctx, dstPath)
			if err == nil && len(existing) > 0 {
				res.Skipped = append(res.Skipped, dstPath)
				continue
			}
		}

		if !m.dryRun {
			if err := m.writer.Write(ctx, dstPath, data); err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("write %q: %w", dstPath, err))
				continue
			}
		}
		res.Mirrored = append(res.Mirrored, dstPath)
	}

	return res, nil
}
