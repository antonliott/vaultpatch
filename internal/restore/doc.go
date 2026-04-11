// Package restore implements snapshot-based secret restoration for Vault namespaces.
//
// It reads a previously saved [snapshot.Snapshot] and writes each secret back
// to the target Vault namespace via a [Writer] interface. Dry-run mode is
// supported so operators can preview which paths would be restored before
// committing any changes.
//
// Typical usage:
//
//	snap, _ := snapshot.Load("backup.json")
//	r := restore.NewRestorer(vaultClient, dryRun)
//	result, _ := r.Apply(ctx, snap)
//	fmt.Println(result.Summary())
package restore
