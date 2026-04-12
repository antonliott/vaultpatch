// Package sync provides functionality to synchronise secrets between
// two Vault namespaces, copying keys from a source to a destination.
package sync

import (
	"context"
	"fmt"
)

// SecretReader reads all secrets at a given path.
type SecretReader interface {
	ListSecrets(ctx context.Context, mount, prefix string) ([]string, error)
	ReadSecret(ctx context.Context, mount, path string) (map[string]interface{}, error)
}

// SecretWriter writes a secret to a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, mount, path string, data map[string]interface{}) error
}

// Options controls the behaviour of a sync operation.
type Options struct {
	Mount  string
	Prefix string
	DryRun bool
}

// Syncer copies secrets from a source namespace to a destination namespace.
type Syncer struct {
	src SecretReader
	dst SecretWriter
}

// NewSyncer creates a new Syncer.
func NewSyncer(src SecretReader, dst SecretWriter) *Syncer {
	return &Syncer{src: src, dst: dst}
}

// Apply performs the sync operation and returns a Result.
func (s *Syncer) Apply(ctx context.Context, opts Options) Result {
	res := Result{DryRun: opts.DryRun}

	keys, err := s.src.ListSecrets(ctx, opts.Mount, opts.Prefix)
	if err != nil {
		res.Errors = append(res.Errors, fmt.Errorf("list %s: %w", opts.Prefix, err))
		return res
	}

	for _, key := range keys {
		data, err := s.src.ReadSecret(ctx, opts.Mount, key)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("read %s: %w", key, err))
			continue
		}

		if !opts.DryRun {
			if err := s.dst.WriteSecret(ctx, opts.Mount, key, data); err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("write %s: %w", key, err))
				continue
			}
		}
		res.Synced = append(res.Synced, key)
	}

	return res
}
