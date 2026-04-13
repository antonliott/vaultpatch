// Package mask provides key-pattern-based redaction of sensitive secret
// values. It is used throughout vaultpatch when secrets are printed to
// stdout, written to audit logs, or included in diff output, ensuring
// that credentials are never accidentally exposed in plain text.
//
// Usage:
//
//	m := mask.NewDefault()
//	safe := m.Apply(rawSecrets)
package mask
