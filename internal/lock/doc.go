// Package lock implements lightweight distributed locking for Vault secret paths.
//
// Locks are stored as KV secrets under a configurable path prefix (e.g. locks/<mount>/<key>).
// Each lock record contains the owner identifier, acquisition timestamp, and TTL.
// A lock is considered stale — and eligible for takeover — once its TTL has elapsed.
//
// Usage:
//
//	l := lock.NewLock(vaultWriter, "locks/secret/myapp", hostname, 60*time.Second)
//	if err := l.Acquire(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer l.Release(ctx)
//
// This package is used internally by the patch and rotate commands to serialise
// concurrent modifications to the same secret path.
package lock
