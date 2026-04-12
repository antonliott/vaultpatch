// Package lock provides distributed locking for Vault secret paths
// to prevent concurrent modifications during patch or rotate operations.
package lock

import (
	"context"
	"fmt"
	"time"
)

// Writer defines the interface for writing and deleting Vault KV entries.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]interface{}) error
	Delete(ctx context.Context, path string) error
	Read(ctx context.Context, path string) (map[string]interface{}, error)
}

// Lock represents a distributed lock stored as a Vault KV secret.
type Lock struct {
	writer    Writer
	lockPath  string
	owner     string
	ttl       time.Duration
	acquiredAt time.Time
}

// NewLock creates a new Lock for the given path and owner identifier.
func NewLock(w Writer, lockPath, owner string, ttl time.Duration) *Lock {
	return &Lock{
		writer:   w,
		lockPath: lockPath,
		owner:    owner,
		ttl:      ttl,
	}
}

// Acquire attempts to acquire the lock. Returns an error if already held.
func (l *Lock) Acquire(ctx context.Context) error {
	existing, err := l.writer.Read(ctx, l.lockPath)
	if err == nil && existing != nil {
		acquiredStr, _ := existing["acquired_at"].(string)
		if acquiredStr != "" {
			t, parseErr := time.Parse(time.RFC3339, acquiredStr)
			if parseErr == nil && time.Since(t) < l.ttl {
				return fmt.Errorf("lock held by %q since %s", existing["owner"], acquiredStr)
			}
		}
	}

	l.acquiredAt = time.Now().UTC()
	data := map[string]interface{}{
		"owner":       l.owner,
		"acquired_at": l.acquiredAt.Format(time.RFC3339),
		"ttl_seconds": int(l.ttl.Seconds()),
	}
	if err := l.writer.Write(ctx, l.lockPath, data); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	return nil
}

// Release deletes the lock secret from Vault.
func (l *Lock) Release(ctx context.Context) error {
	if err := l.writer.Delete(ctx, l.lockPath); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}
	return nil
}

// IsExpired reports whether the lock TTL has elapsed since acquisition.
func (l *Lock) IsExpired() bool {
	if l.acquiredAt.IsZero() {
		return true
	}
	return time.Since(l.acquiredAt) >= l.ttl
}
