package patch

import (
	"context"
	"fmt"

	"github.com/youorg/vaultpatch/internal/diff"
)

// Applier applies a set of diff changes to a target Vault path.
type Applier struct {
	writer SecretWriter
	dryRun bool
}

// SecretWriter is the interface for writing secrets to Vault.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	DeleteSecret(ctx context.Context, path string) error
}

// NewApplier creates a new Applier.
func NewApplier(w SecretWriter, dryRun bool) *Applier {
	return &Applier{writer: w, dryRun: dryRun}
}

// Apply applies the given diff changes to the target path.
// If dryRun is true, changes are logged but not persisted.
func (a *Applier) Apply(ctx context.Context, path string, changes []diff.Change) (int, error) {
	if len(changes) == 0 {
		return 0, nil
	}

	updates := make(map[string]interface{})
	var toDelete []string

	for _, c := range changes {
		switch c.Type {
		case diff.Added, diff.Changed:
			updates[c.Key] = c.NewValue
		case diff.Removed:
			toDelete = append(toDelete, c.Key)
		}
	}

	if a.dryRun {
		fmt.Printf("[dry-run] would write %d key(s) to %s\n", len(updates), path)
		fmt.Printf("[dry-run] would delete %d key(s) from %s\n", len(toDelete), path)
		return len(changes), nil
	}

	if len(updates) > 0 {
		if err := a.writer.WriteSecret(ctx, path, updates); err != nil {
			return 0, fmt.Errorf("writing secrets to %s: %w", path, err)
		}
	}

	for _, key := range toDelete {
		if err := a.writer.DeleteSecret(ctx, path+"/"+key, nil); err != nil {
			return 0, fmt.Errorf("deleting key %s from %s: %w", key, path, err)
		}
	}

	return len(changes), nil
}
