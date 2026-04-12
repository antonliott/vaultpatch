// Package watch provides secret change detection by polling Vault paths
// at a configurable interval and emitting diffs when changes are detected.
package watch

import (
	"context"
	"time"

	"github.com/your-org/vaultpatch/internal/diff"
)

// SecretReader reads a map of key/value secrets from a given path.
type SecretReader interface {
	ReadSecrets(ctx context.Context, path string) (map[string]string, error)
}

// Event is emitted when a change is detected on a watched path.
type Event struct {
	Path    string
	Deltas  []diff.Delta
	SeenAt  time.Time
}

// Watcher polls a Vault path and sends Events on the Changes channel.
type Watcher struct {
	reader   SecretReader
	path     string
	interval time.Duration
	Changes  chan Event
}

// NewWatcher creates a Watcher for the given path and poll interval.
// interval must be positive; if zero or negative it defaults to 30s.
func NewWatcher(reader SecretReader, path string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Watcher{
		reader:   reader,
		path:     path,
		interval: interval,
		Changes:  make(chan Event, 8),
	}
}

// Run starts polling until ctx is cancelled. It performs an initial read
// immediately, then emits an Event whenever the secrets differ from the
// previous snapshot.
func (w *Watcher) Run(ctx context.Context) error {
	prev, err := w.reader.ReadSecrets(ctx, w.path)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(w.Changes)
			return ctx.Err()
		case <-ticker.C:
			curr, err := w.reader.ReadSecrets(ctx, w.path)
			if err != nil {
				// non-fatal: skip this tick, keep previous state
				continue
			}
			deltas := diff.Compare(prev, curr)
			if len(deltas) > 0 {
				w.Changes <- Event{
					Path:   w.path,
					Deltas: deltas,
					SeenAt: time.Now().UTC(),
				}
				prev = curr
			}
		}
	}
}
