// Package archive implements soft-deletion of Vault secrets by copying them
// to a configurable archive path prefix before removal. Each archived entry
// is annotated with an _archived_at RFC-3339 timestamp so operators can
// audit when a secret was retired.
//
// Typical usage:
//
//	a := archive.NewArchiver(vaultReader, vaultWriter, "secret/archive", dryRun)
//	res, err := a.Apply(ctx, []string{"secret/prod/db", "secret/prod/api"})
package archive
