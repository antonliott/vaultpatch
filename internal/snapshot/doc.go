// Package snapshot provides functionality for capturing, persisting,
// and loading point-in-time snapshots of HashiCorp Vault secrets.
//
// A Snapshot records the namespace, capture timestamp, and a flat
// map of secret paths to their string values. Snapshots can be saved
// to and loaded from JSON files on disk, enabling offline diffing
// between two Vault namespaces or across time.
//
// Usage:
//
//	collector := snapshot.NewCollector("prod", vaultClient)
//	snap, err := collector.Collect(ctx, "secret/")
//	if err != nil { ... }
//
//	if err := snapshot.Save(snap, "prod-snapshot.json"); err != nil { ... }
//
//	loaded, err := snapshot.Load("prod-snapshot.json")
package snapshot
