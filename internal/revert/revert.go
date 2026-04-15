// Package revert provides functionality to revert a secret path to a
// previous KV v2 version in HashiCorp Vault.
package revert

import (
	"context"
	"fmt"
)

// Reader reads secret data at a given path and version.
type Reader interface {
	ReadVersion(ctx context.Context, path string, version int) (map[string]string, error)
	Write(ctx context.Context, path string, data map[string]string) error
}

// Options controls revert behaviour.
type Options struct {
	Path    string
	Version int
	DryRun  bool
}

// Result holds the outcome of a revert operation.
type Result struct {
	Path        string
	Version     int
	Data        map[string]string
	DryRun      bool
	Err         error
}

// Reverter reverts a secret to a previous version.
type Reverter struct {
	reader Reader
}

// NewReverter creates a new Reverter backed by the given Reader.
func NewReverter(r Reader) *Reverter {
	return &Reverter{reader: r}
}

// Apply reads the specified version of a secret and, unless DryRun is set,
// writes it back as the current version.
func (rv *Reverter) Apply(ctx context.Context, opts Options) Result {
	if opts.Version < 1 {
		return Result{
			Path:    opts.Path,
			Version: opts.Version,
			DryRun:  opts.DryRun,
			Err:     fmt.Errorf("version must be >= 1, got %d", opts.Version),
		}
	}

	data, err := rv.reader.ReadVersion(ctx, opts.Path, opts.Version)
	if err != nil {
		return Result{
			Path:    opts.Path,
			Version: opts.Version,
			DryRun:  opts.DryRun,
			Err:     fmt.Errorf("read version %d of %q: %w", opts.Version, opts.Path, err),
		}
	}

	if opts.DryRun {
		return Result{Path: opts.Path, Version: opts.Version, Data: data, DryRun: true}
	}

	if err := rv.reader.Write(ctx, opts.Path, data); err != nil {
		return Result{
			Path:    opts.Path,
			Version: opts.Version,
			Data:    data,
			Err:     fmt.Errorf("write revert of %q to version %d: %w", opts.Path, opts.Version, err),
		}
	}

	return Result{Path: opts.Path, Version: opts.Version, Data: data}
}
