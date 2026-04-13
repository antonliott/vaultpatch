// Package trim provides functionality to remove stale or expired secrets
// from a Vault namespace based on a configurable age threshold.
package trim

import (
	"context"
	"fmt"
	"time"
)

// SecretLister lists secret keys under a given path.
type SecretLister interface {
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// SecretDeleter deletes a secret at a given path.
type SecretDeleter interface {
	DeleteSecret(ctx context.Context, path string) error
}

// SecretReader reads metadata (including created_time) for a secret.
type SecretReader interface {
	ReadSecretMetadata(ctx context.Context, path string) (map[string]interface{}, error)
}

// Client combines all required vault operations for trimming.
type Client interface {
	SecretLister
	SecretDeleter
	SecretReader
}

// Result holds the outcome of a trim operation.
type Result struct {
	Deleted []string
	Skipped []string
	Errors  []error
	DryRun  bool
}

// HasErrors returns true if any errors were encountered.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a human-readable summary of the trim result.
func (r *Result) Summary() string {
	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}
	return fmt.Sprintf("[%s] deleted=%d skipped=%d errors=%d",
		mode, len(r.Deleted), len(r.Skipped), len(r.Errors))
}

// Trimmer removes secrets older than a given threshold.
type Trimmer struct {
	client    Client
	maxAge    time.Duration
	dryRun    bool
}

// NewTrimmer creates a new Trimmer.
func NewTrimmer(client Client, maxAge time.Duration, dryRun bool) *Trimmer {
	return &Trimmer{client: client, maxAge: maxAge, dryRun: dryRun}
}

// Apply lists secrets under path and deletes those older than maxAge.
func (t *Trimmer) Apply(ctx context.Context, path string) (*Result, error) {
	keys, err := t.client.ListSecrets(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	result := &Result{DryRun: t.dryRun}
	cutoff := time.Now().UTC().Add(-t.maxAge)

	for _, key := range keys {
		full := path + "/" + key
		meta, err := t.client.ReadSecretMetadata(ctx, full)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("read metadata %s: %w", full, err))
			continue
		}

		createdRaw, ok := meta["created_time"].(string)
		if !ok {
			result.Skipped = append(result.Skipped, full)
			continue
		}

		created, err := time.Parse(time.RFC3339Nano, createdRaw)
		if err != nil {
			result.Skipped = append(result.Skipped, full)
			continue
		}

		if created.Before(cutoff) {
			if !t.dryRun {
				if err := t.client.DeleteSecret(ctx, full); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("delete %s: %w", full, err))
					continue
				}
			}
			result.Deleted = append(result.Deleted, full)
		} else {
			result.Skipped = append(result.Skipped, full)
		}
	}

	return result, nil
}
