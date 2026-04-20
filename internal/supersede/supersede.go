// Package supersede provides functionality to override secrets at a target
// path with values from a source path, optionally restricted to a set of keys.
package supersede

import (
	"context"
	"fmt"
)

// Reader reads secrets from a Vault path.
type Reader interface {
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

// Options configures a Superseder operation.
type Options struct {
	// Keys restricts which keys are superseded. If empty, all keys from source
	// are applied.
	Keys []string
	// DryRun reports what would change without writing.
	DryRun bool
}

// Delta records a single key override.
type Delta struct {
	Key      string
	OldValue string
	NewValue string
}

// Result summarises a supersede operation.
type Result struct {
	Deltas []Delta
	DryRun bool
}

// Applied returns the number of keys that were (or would be) overridden.
func (r Result) Applied() int { return len(r.Deltas) }

// Superseder applies source secret values onto a target path.
type Superseder struct {
	rw   ReadWriter
	opts Options
}

// New creates a new Superseder.
func New(rw ReadWriter, opts Options) *Superseder {
	return &Superseder{rw: rw, opts: opts}
}

// Apply reads src, merges selected keys into dst, and writes the result.
func (s *Superseder) Apply(ctx context.Context, src, dst string) (Result, error) {
	srcData, err := s.rw.Read(ctx, src)
	if err != nil {
		return Result{}, fmt.Errorf("supersede: read source %q: %w", src, err)
	}

	dstData, err := s.rw.Read(ctx, dst)
	if err != nil {
		return Result{}, fmt.Errorf("supersede: read target %q: %w", dst, err)
	}

	keys := s.opts.Keys
	if len(keys) == 0 {
		for k := range srcData {
			keys = append(keys, k)
		}
	}

	var deltas []Delta
	merged := make(map[string]string, len(dstData))
	for k, v := range dstData {
		merged[k] = v
	}

	for _, k := range keys {
		newVal, ok := srcData[k]
		if !ok {
			continue
		}
		oldVal := merged[k]
		if oldVal == newVal {
			continue
		}
		deltas = append(deltas, Delta{Key: k, OldValue: oldVal, NewValue: newVal})
		merged[k] = newVal
	}

	res := Result{Deltas: deltas, DryRun: s.opts.DryRun}
	if s.opts.DryRun || len(deltas) == 0 {
		return res, nil
	}

	if err := s.rw.Write(ctx, dst, merged); err != nil {
		return res, fmt.Errorf("supersede: write target %q: %w", dst, err)
	}
	return res, nil
}
