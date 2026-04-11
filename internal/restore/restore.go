// Package restore provides functionality to restore Vault secrets from a snapshot.
package restore

import (
	"context"
	"fmt"

	"github.com/example/vaultpatch/internal/snapshot"
)

// Writer defines the interface for writing a secret to Vault.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]interface{}) error
}

// Restorer applies a snapshot's secrets back to a Vault namespace.
type Restorer struct {
	writer  Writer
	dryRun  bool
}

// NewRestorer creates a new Restorer.
func NewRestorer(w Writer, dryRun bool) *Restorer {
	return &Restorer{writer: w, dryRun: dryRun}
}

// RestoreResult holds the outcome of a restore operation.
type RestoreResult struct {
	Restored []string
	Skipped  []string
	Errors   []error
	DryRun   bool
}

// HasErrors returns true if any errors occurred during restore.
func (r *RestoreResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a human-readable summary of the restore operation.
func (r *RestoreResult) Summary() string {
	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}
	return fmt.Sprintf("[%s] restored=%d skipped=%d errors=%d",
		mode, len(r.Restored), len(r.Skipped), len(r.Errors))
}

// Apply writes all secrets from the snapshot to Vault.
func (r *Restorer) Apply(ctx context.Context, snap *snapshot.Snapshot) (*RestoreResult, error) {
	result := &RestoreResult{DryRun: r.dryRun}

	for path, data := range snap.Secrets {
		if r.dryRun {
			result.Skipped = append(result.Skipped, path)
			continue
		}
		if err := r.writer.Write(ctx, path, data); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("write %s: %w", path, err))
			continue
		}
		result.Restored = append(result.Restored, path)
	}

	return result, nil
}
