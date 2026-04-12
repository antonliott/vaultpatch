package watch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/watch"
)

type mockReader struct {
	calls   int
	results []map[string]string
	err     error
}

func (m *mockReader) ReadSecrets(_ context.Context, _ string) (map[string]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	idx := m.calls
	if idx >= len(m.results) {
		idx = len(m.results) - 1
	}
	m.calls++
	return m.results[idx], nil
}

func TestNewWatcher_DefaultInterval(t *testing.T) {
	r := &mockReader{results: []map[string]string{{}}}
	w := watch.NewWatcher(r, "secret/app", 0)
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestRun_EmitsEventOnChange(t *testing.T) {
	r := &mockReader{
		results: []map[string]string{
			{"key": "v1"},
			{"key": "v2"},
		},
	}
	w := watch.NewWatcher(r, "secret/app", 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go w.Run(ctx) //nolint:errcheck

	select {
	case ev := <-w.Changes:
		if ev.Path != "secret/app" {
			t.Errorf("unexpected path: %s", ev.Path)
		}
		if len(ev.Deltas) == 0 {
			t.Error("expected at least one delta")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for change event")
	}
}

func TestRun_NoEventWhenUnchanged(t *testing.T) {
	r := &mockReader{
		results: []map[string]string{
			{"key": "same"},
		},
	}
	w := watch.NewWatcher(r, "secret/app", 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	go w.Run(ctx) //nolint:errcheck

	select {
	case ev := <-w.Changes:
		t.Errorf("unexpected event: %+v", ev)
	case <-ctx.Done():
		// expected: no changes
	}
}

func TestRun_SkipTickOnReadError(t *testing.T) {
	call := 0
	r := &mockReader{}
	// first call succeeds, second returns error, third has a change
	results := []map[string]string{{"k": "1"}, {"k": "2"}}
	errs := []error{nil, errors.New("transient"), nil}

	customReader := &callbackReader{fn: func() (map[string]string, error) {
		i := call
		if i >= len(errs) {
			i = len(errs) - 1
		}
		call++
		if errs[i] != nil {
			return nil, errs[i]
		}
		j := i
		if j >= len(results) {
			j = len(results) - 1
		}
		return results[j], nil
	}}
	_ = r

	w := watch.NewWatcher(customReader, "secret/app", 10*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go w.Run(ctx) //nolint:errcheck

	select {
	case <-w.Changes:
		// received at least one change, transient error was skipped
	case <-ctx.Done():
		t.Fatal("timed out — transient error may have broken the loop")
	}
}

type callbackReader struct {
	fn func() (map[string]string, error)
}

func (c *callbackReader) ReadSecrets(_ context.Context, _ string) (map[string]string, error) {
	return c.fn()
}
