// Package redact provides a Redactor type that replaces sensitive secret
// values with a [REDACTED] placeholder before the data is surfaced in CLI
// output, audit logs, or exported files.
//
// Usage:
//
//	r := redact.NewDefault()
//	safe := r.Apply(secrets)
//
// A custom key list can be supplied via redact.New(keys).
package redact
