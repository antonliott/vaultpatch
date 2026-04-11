// Package audit provides structured, append-only audit logging for
// vaultpatch operations.
//
// Each operation (diff, apply, dry-run) that mutates or inspects Vault
// secrets should produce an [Event] via [Logger.Log] or the convenience
// helpers such as [Logger.LogApply].
//
// Events are written as newline-delimited JSON (NDJSON) to any io.Writer,
// making them easy to ship to log aggregators or persist to a local file.
//
// Example usage:
//
//	f, _ := os.OpenFile("vaultpatch-audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
//	logger := audit.NewLogger(f)
//	logger.LogApply("prod", "secret/db", []string{"password"}, false, nil)
package audit
