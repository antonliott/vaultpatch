// Package compare implements cross-namespace secret diffing for vaultpatch.
//
// It lists secrets from a source Vault namespace, reads corresponding entries
// from a target namespace, and returns a structured Result describing every
// key-level addition, removal, or change. The Format function renders the
// result as a human-readable unified-style diff suitable for CLI output.
package compare
