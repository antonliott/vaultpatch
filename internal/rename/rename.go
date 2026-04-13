// Package rename provides functionality to rename secret keys
// across one or more Vault KV paths.
package rename

import (
	"context"
	"fmt"
)

// Reader reads a secret at the given path.
type Reader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes a secret at the given path.
type Writer interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Result holds the outcome of a rename operation.
type Result struct {
	Path    string
	OldKey  string
	NewKey  string
	DryRun  bool
	Skipped bool
	Err     error
}

// Renamer renames a key within Vault KV secrets.
type Renamer struct {
	client ReadWriter
}

// NewRenamer creates a new Renamer backed by the given ReadWriter.
func NewRenamer(rw ReadWriter) *Renamer {
	return &Renamer{client: rw}
}

// Apply renames oldKey to newKey in the secret at path.
// If dryRun is true, no writes are performed.
func (r *Renamer) Apply(ctx context.Context, path, oldKey, newKey string, dryRun bool) Result {
	res := Result{Path: path, OldKey: oldKey, NewKey: newKey, DryRun: dryRun}

	data, err := r.client.ReadSecret(ctx, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		return res
	}

	val, ok := data[oldKey]
	if !ok {
		res.Skipped = true
		return res
	}

	if _, exists := data[newKey]; exists {
		res.Err = fmt.Errorf("key %q already exists in %s", newKey, path)
		return res
	}

	if dryRun {
		return res
	}

	updated := make(map[string]string, len(data))
	for k, v := range data {
		updated[k] = v
	}
	updated[newKey] = val
	delete(updated, oldKey)

	if err := r.client.WriteSecret(ctx, path, updated); err != nil {
		res.Err = fmt.Errorf("write %s: %w", path, err)
	}
	return res
}
