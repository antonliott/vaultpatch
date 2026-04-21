// Package annotate provides functionality for adding, updating, and removing
// metadata annotations on Vault secret paths.
package annotate

import (
	"context"
	"fmt"
)

// Writer can read and write secret data at a given path.
type Writer interface {
	Read(ctx context.Context, path string) (map[string]string, error)
	Write(ctx context.Context, path string, data map[string]string) error
}

// Options configures the annotation operation.
type Options struct {
	Path        string
	Annotations map[string]string
	RemoveKeys  []string
	DryRun      bool
}

// Result holds the outcome of an annotation operation.
type Result struct {
	Path    string
	Added   int
	Updated int
	Removed int
	DryRun  bool
	Err     error
}

// Annotator applies annotation changes to a secret path.
type Annotator struct {
	w Writer
}

// NewAnnotator creates a new Annotator backed by the given Writer.
func NewAnnotator(w Writer) *Annotator {
	return &Annotator{w: w}
}

// Apply reads the secret at opts.Path, merges annotations, removes specified
// keys, and writes the result back unless DryRun is set.
func (a *Annotator) Apply(ctx context.Context, opts Options) Result {
	res := Result{Path: opts.Path, DryRun: opts.DryRun}

	current, err := a.w.Read(ctx, opts.Path)
	if err != nil {
		res.Err = fmt.Errorf("annotate: read %s: %w", opts.Path, err)
		return res
	}

	updated := make(map[string]string, len(current))
	for k, v := range current {
		updated[k] = v
	}

	for k, v := range opts.Annotations {
		if _, exists := updated[k]; exists {
			res.Updated++
		} else {
			res.Added++
		}
		updated[k] = v
	}

	removeSet := make(map[string]struct{}, len(opts.RemoveKeys))
	for _, k := range opts.RemoveKeys {
		removeSet[k] = struct{}{}
	}
	for k := range removeSet {
		if _, exists := updated[k]; exists {
			delete(updated, k)
			res.Removed++
		}
	}

	if opts.DryRun {
		return res
	}

	if err := a.w.Write(ctx, opts.Path, updated); err != nil {
		res.Err = fmt.Errorf("annotate: write %s: %w", opts.Path, err)
	}
	return res
}
