// Package rollback implements secret rollback for the vaultpatch CLI.
//
// A rollback operation reads a previously saved [snapshot.Snapshot] and
// reverts each secret path recorded in that snapshot to its captured value.
// Paths whose current value already matches the snapshot are skipped.
//
// Dry-run mode reports which paths would be reverted without writing to Vault.
package rollback
