// Package inherit implements secret inheritance for vaultpatch.
//
// It reads key-value pairs from a designated parent Vault path and propagates
// them to one or more child paths. Existing keys in child paths are preserved
// unless the Overwrite option is set. Dry-run mode reports changes without
// writing to Vault.
//
// Typical usage:
//
//	inheritor := inherit.NewInheritor(vaultRW, inherit.Options{DryRun: false})
//	result, err := inheritor.Apply(ctx, "secret/base", []string{"secret/svc-a", "secret/svc-b"})
package inherit
