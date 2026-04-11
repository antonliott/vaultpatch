// Package rotate implements secret rotation across HashiCorp Vault paths.
//
// It supports merging new key values into existing secrets while preserving
// untouched keys, and provides a dry-run mode that reports what would change
// without performing any writes.
//
// Usage:
//
//	w := vaultClient // implements rotate.Writer
//	r := rotate.NewRotator(w, dryRun)
//	results := r.Apply(ctx, []rotate.RotateRequest{
//	    {Path: "secret/db", Updates: map[string]string{"password": newPassword}},
//	})
package rotate
