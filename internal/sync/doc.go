// Package sync implements cross-namespace secret synchronisation for
// vaultpatch. It reads secrets from a source Vault namespace and writes
// them to a destination namespace, supporting dry-run mode so that
// operators can preview changes before they are applied.
//
// Usage:
//
//	syncer := sync.NewSyncer(srcClient, dstClient)
//	result := syncer.Apply(ctx, sync.Options{
//		Mount:  "secret",
//		Prefix: "app/",
//		DryRun: false,
//	})
//	fmt.Println(result.Summary())
package sync
