// Package audit provides structured logging of patch operations
// applied across Vault namespaces.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// EventType classifies the kind of audit event.
type EventType string

const (
	EventApply  EventType = "apply"
	EventDiff   EventType = "diff"
	EventDryRun EventType = "dry_run"
)

// Event represents a single auditable action performed by vaultpatch.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Namespace string    `json:"namespace"`
	Path      string    `json:"path"`
	Keys      []string  `json:"keys,omitempty"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit events as newline-delimited JSON to an io.Writer.
type Logger struct {
	w io.Writer
}

// NewLogger creates a new Logger that writes to w.
func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Log records an audit event.
func (l *Logger) Log(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("audit: marshal event: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	if err != nil {
		return fmt.Errorf("audit: write event: %w", err)
	}
	return nil
}

// LogApply is a convenience method for logging an apply operation.
func (l *Logger) LogApply(namespace, path string, keys []string, dryRun bool, err error) error {
	et := EventApply
	if dryRun {
		et = EventDryRun
	}
	ev := Event{
		Type:      et,
		Namespace: namespace,
		Path:      path,
		Keys:      keys,
		Success:   err == nil,
	}
	if err != nil {
		ev.Message = err.Error()
	}
	return l.Log(ev)
}
