// Package prefix provides utilities for adding, removing, or replacing
// key prefixes across secrets stored in a Vault KV path.
package prefix

import (
	"context"
	"fmt"
	"strings"
)

// SecretReader reads a map of key/value pairs from a given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// SecretWriter writes a map of key/value pairs to a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// Options controls how prefix transformation is applied.
type Options struct {
	OldPrefix string
	NewPrefix string
	DryRun    bool
}

// Result holds the outcome of a prefix operation.
type Result struct {
	Path    string
	Renamed map[string]string // oldKey -> newKey
	Skipped []string
	DryRun  bool
	Err     error
}

// Renamer applies prefix transformations to secrets at a path.
type Renamer struct {
	rw SecretReaderWriter
}

// SecretReaderWriter combines reading and writing.
type SecretReaderWriter interface {
	SecretReader
	SecretWriter
}

// NewRenamer constructs a new prefix Renamer.
func NewRenamer(rw SecretReaderWriter) *Renamer {
	return &Renamer{rw: rw}
}

// Apply reads secrets at path and renames keys matching OldPrefix to NewPrefix.
func (r *Renamer) Apply(ctx context.Context, path string, opts Options) Result {
	result := Result{
		Path:    path,
		Renamed: make(map[string]string),
		DryRun:  opts.DryRun,
	}

	secrets, err := r.rw.ReadSecret(ctx, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}

	updated := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if strings.HasPrefix(k, opts.OldPrefix) {
			newKey := opts.NewPrefix + strings.TrimPrefix(k, opts.OldPrefix)
			result.Renamed[k] = newKey
			updated[newKey] = v
		} else {
			result.Skipped = append(result.Skipped, k)
			updated[k] = v
		}
	}

	if opts.DryRun || len(result.Renamed) == 0 {
		return result
	}

	if err := r.rw.WriteSecret(ctx, path, updated); err != nil {
		result.Err = fmt.Errorf("write %s: %w", path, err)
	}
	return result
}
