// Package compact removes duplicate or redundant key-value pairs across
// multiple Vault secret paths, keeping only the first occurrence.
package compact

import (
	"context"
	"fmt"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secrets to a Vault path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Result holds the outcome of a compact operation.
type Result struct {
	Path    string
	Removed []string
	DryRun  bool
	Err     error
}

// Compactor removes keys from paths that already appeared in earlier paths.
type Compactor struct {
	rw     ReadWriter
	dryRun bool
}

// New creates a new Compactor.
func New(rw ReadWriter, dryRun bool) *Compactor {
	return &Compactor{rw: rw, dryRun: dryRun}
}

// Apply scans paths in order and removes keys already seen in a prior path.
func (c *Compactor) Apply(ctx context.Context, paths []string) ([]Result, error) {
	seen := make(map[string]struct{})
	var results []Result

	for _, path := range paths {
		data, err := c.rw.Read(ctx, path)
		if err != nil {
			results = append(results, Result{Path: path, Err: fmt.Errorf("read: %w", err)})
			continue
		}

		var removed []string
		compacted := make(map[string]string, len(data))

		for k, v := range data {
			if _, exists := seen[k]; exists {
				removed = append(removed, k)
			} else {
				seen[k] = struct{}{}
				compacted[k] = v
			}
		}

		if len(removed) > 0 && !c.dryRun {
			if err := c.rw.Write(ctx, path, compacted); err != nil {
				results = append(results, Result{Path: path, Removed: removed, Err: fmt.Errorf("write: %w", err)})
				continue
			}
		}

		results = append(results, Result{Path: path, Removed: removed, DryRun: c.dryRun})
	}

	return results, nil
}
