// Package rollback provides functionality to revert secrets in a Vault
// namespace to a previously captured snapshot state.
package rollback

import (
	"context"
	"fmt"

	"github.com/your-org/vaultpatch/internal/snapshot"
)

// Writer is the interface for writing a secret back to Vault.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]interface{}) error
	Read(ctx context.Context, path string) (map[string]interface{}, error)
}

// Rollbacker reverts secrets to a prior snapshot.
type Rollbacker struct {
	writer  Writer
	dryRun  bool
}

// NewRollbacker creates a new Rollbacker.
func NewRollbacker(w Writer, dryRun bool) *Rollbacker {
	return &Rollbacker{writer: w, dryRun: dryRun}
}

// Apply reverts each secret in the snapshot to its recorded value.
func (r *Rollbacker) Apply(ctx context.Context, snap *snapshot.Snapshot) (*Result, error) {
	res := &Result{DryRun: r.dryRun}

	for path, data := range snap.Secrets {
		current, err := r.writer.Read(ctx, path)
		if err != nil {
			res.addError(path, fmt.Errorf("read current: %w", err))
			continue
		}

		if secretsEqual(current, data) {
			res.Skipped++
			continue
		}

		if r.dryRun {
			res.WouldRevert = append(res.WouldRevert, path)
			continue
		}

		if err := r.writer.Write(ctx, path, data); err != nil {
			res.addError(path, fmt.Errorf("write revert: %w", err))
			continue
		}
		res.Reverted = append(res.Reverted, path)
	}

	return res, nil
}

func secretsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || fmt.Sprintf("%v", va) != fmt.Sprintf("%v", vb) {
			return false
		}
	}
	return true
}
