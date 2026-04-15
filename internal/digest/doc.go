// Package digest provides content-addressable hashing for Vault secret paths.
//
// It computes a stable SHA-256 digest for each secret by hashing the sorted
// key=value pairs of the secret's data. This allows callers to detect which
// paths have changed between two snapshots or environments without storing
// or transmitting raw secret values.
//
// Typical usage:
//
//	res, err := digest.Compute(vaultReader, "secret/myapp")
//	changed := digest.Changed(baseline, res)
package digest
