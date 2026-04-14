// Package rename implements key-rename operations for Vault KV secrets.
//
// It reads a secret at a given path, renames the specified key to a new name,
// and writes the updated map back. The operation is idempotent when the source
// key is absent (the result is marked as skipped) and guards against
// overwriting an existing key with the same target name.
//
// Dry-run mode is supported: when enabled the rename is validated but no
// write is issued to Vault.
//
// # Error conditions
//
// The package surfaces distinct sentinel errors for common failure modes:
//
//   - ErrKeyNotFound – the source key does not exist in the secret.
//   - ErrKeyConflict  – the target key already exists and would be overwritten.
//   - ErrPathNotFound – the secret path does not exist in Vault.
//
// Callers can use errors.Is to distinguish these cases and decide whether to
// treat them as fatal or as expected no-ops in their workflow.
package rename
