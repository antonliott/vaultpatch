// Package rotate provides utilities for rotating secrets across Vault paths.
package rotate

import (
	"context"
	"fmt"
)

// Writer defines the interface for writing a secret to a Vault path.
type Writer interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// Rotator rotates secrets by replacing specified keys with new values.
type Rotator struct {
	writer Writer
	dryRun bool
}

// NewRotator creates a new Rotator.
func NewRotator(w Writer, dryRun bool) *Rotator {
	return &Rotator{writer: w, dryRun: dryRun}
}

// RotateRequest describes a single rotation operation.
type RotateRequest struct {
	Path    string
	Updates map[string]string
}

// RotateResult holds the outcome of a rotation.
type RotateResult struct {
	Path    string
	DryRun  bool
	Err     error
}

// Apply performs the rotation for the given requests.
func (r *Rotator) Apply(ctx context.Context, reqs []RotateRequest) []RotateResult {
	results := make([]RotateResult, 0, len(reqs))
	for _, req := range reqs {
		res := RotateResult{Path: req.Path, DryRun: r.dryRun}
		if !r.dryRun {
			existing, err := r.writer.ReadSecret(ctx, req.Path)
			if err != nil {
				res.Err = fmt.Errorf("read %s: %w", req.Path, err)
				results = append(results, res)
				continue
			}
			merged := make(map[string]interface{}, len(existing))
			for k, v := range existing {
				merged[k] = v
			}
			for k, v := range req.Updates {
				merged[k] = v
			}
			if err := r.writer.WriteSecret(ctx, req.Path, merged); err != nil {
				res.Err = fmt.Errorf("write %s: %w", req.Path, err)
			}
		}
		results = append(results, res)
	}
	return results
}
