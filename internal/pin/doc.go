// Package pin provides secret pinning and drift detection for Vault paths.
//
// A PinEntry captures the key-value state of a secret at a point in time.
// CheckDrift compares a live secret against its pinned state and reports
// any additions, removals, or value changes. Restore rewrites the pinned
// values back to Vault, optionally in dry-run mode.
package pin
