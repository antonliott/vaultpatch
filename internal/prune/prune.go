// Package prune removes secret keys from Vault paths that match a given filter.
package prune

import (
	"context"
	"fmt"
)

// SecretReader reads a secret at the given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// SecretWriter writes a secret to the given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// ReadWriter combines reading and writing.
type ReadWriter interface {
	SecretReader
	SecretWriter
}

// Result holds the outcome of a prune operation.
type Result struct {
	Path        string
	PrunedKeys  []string
	DryRun      bool
	Err         error
}

// Pruner removes matching keys from secrets.
type Pruner struct {
	client ReadWriter
	keys   []string
	dryRun bool
}

// NewPruner creates a Pruner that will remove the given keys.
func NewPruner(client ReadWriter, keys []string, dryRun bool) *Pruner {
	return &Pruner{client: client, keys: keys, dryRun: dryRun}
}

// Apply reads the secret at path, removes matching keys, and writes it back.
func (p *Pruner) Apply(ctx context.Context, path string) Result {
	res := Result{Path: path, DryRun: p.dryRun}

	data, err := p.client.ReadSecret(ctx, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		return res
	}

	updated := make(map[string]interface{}, len(data))
	for k, v := range data {
		updated[k] = v
	}

	for _, key := range p.keys {
		if _, exists := updated[key]; exists {
			res.PrunedKeys = append(res.PrunedKeys, key)
			delete(updated, key)
		}
	}

	if len(res.PrunedKeys) == 0 {
		return res
	}

	if !p.dryRun {
		if err := p.client.WriteSecret(ctx, path, updated); err != nil {
			res.Err = fmt.Errorf("write %s: %w", path, err)
		}
	}

	return res
}
