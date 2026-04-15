// Package inherit provides functionality to propagate secrets from a parent
// path to one or more child paths in Vault, merging keys without overwriting
// existing values unless explicitly requested.
package inherit

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

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Options controls inheritance behaviour.
type Options struct {
	DryRun    bool
	Overwrite bool
}

// Result holds the outcome of an inherit operation.
type Result struct {
	Parent   string
	Children []ChildResult
}

// ChildResult holds the outcome for a single child path.
type ChildResult struct {
	Path    string
	Applied int
	Skipped int
	Err     error
}

// Inheritor propagates secrets from a parent path to child paths.
type Inheritor struct {
	rw  ReadWriter
	opts Options
}

// NewInheritor creates a new Inheritor.
func NewInheritor(rw ReadWriter, opts Options) *Inheritor {
	return &Inheritor{rw: rw, opts: opts}
}

// Apply reads the parent path and merges its keys into each child path.
func (i *Inheritor) Apply(ctx context.Context, parent string, children []string) (Result, error) {
	result := Result{Parent: parent}

	parentData, err := i.rw.ReadSecrets(ctx, parent)
	if err != nil {
		return result, fmt.Errorf("inherit: read parent %q: %w", parent, err)
	}

	for _, child := range children {
		cr := i.applyChild(ctx, parent, child, parentData)
		result.Children = append(result.Children, cr)
	}

	return result, nil
}

func (i *Inheritor) applyChild(ctx context.Context, parent, child string, parentData map[string]string) ChildResult {
	cr := ChildResult{Path: child}

	childData, err := i.rw.ReadSecrets(ctx, child)
	if err != nil {
		cr.Err = fmt.Errorf("inherit: read child %q: %w", child, err)
		return cr
	}

	merged := make(map[string]string, len(childData))
	for k, v := range childData {
		merged[k] = v
	}

	for k, v := range parentData {
		if _, exists := merged[k]; exists && !i.opts.Overwrite {
			cr.Skipped++
			continue
		}
		merged[k] = v
		cr.Applied++
	}

	if cr.Applied == 0 || i.opts.DryRun {
		return cr
	}

	if err := i.rw.WriteSecrets(ctx, child, merged); err != nil {
		cr.Err = fmt.Errorf("inherit: write child %q: %w", child, err)
	}

	return cr
}
