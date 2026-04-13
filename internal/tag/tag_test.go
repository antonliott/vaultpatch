package tag_test

import (
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/tag"
)

// mockWriter implements tag.Writer.
type mockWriter struct {
	called bool
	err    error
	last   tag.Tags
}

func (m *mockWriter) WriteMetadata(_ string, tags tag.Tags) error {
	m.called = true
	m.last = tags
	return m.err
}

func TestCompare_NoChanges(t *testing.T) {
	current := tag.Tags{"env": "prod", "team": "platform"}
	desired := tag.Tags{"env": "prod", "team": "platform"}
	deltas := tag.Compare(current, desired)
	if len(deltas) != 0 {
		t.Fatalf("expected 0 deltas, got %d", len(deltas))
	}
}

func TestCompare_Added(t *testing.T) {
	current := tag.Tags{}
	desired := tag.Tags{"env": "staging"}
	deltas := tag.Compare(current, desired)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Key != "env" || deltas[0].New != "staging" {
		t.Errorf("unexpected delta: %+v", deltas[0])
	}
}

func TestCompare_Changed(t *testing.T) {
	current := tag.Tags{"env": "dev"}
	desired := tag.Tags{"env": "prod"}
	deltas := tag.Compare(current, desired)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Old != "dev" || deltas[0].New != "prod" {
		t.Errorf("unexpected delta: %+v", deltas[0])
	}
}

func TestApply_DryRun(t *testing.T) {
	w := &mockWriter{}
	a := tag.NewApplier(w, true)
	res := a.Apply("secret/app", tag.Tags{}, tag.Tags{"env": "prod"})
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if res.Applied != 1 {
		t.Errorf("expected Applied=1, got %d", res.Applied)
	}
	if w.called {
		t.Error("writer should not be called in dry-run mode")
	}
}

func TestApply_Success(t *testing.T) {
	w := &mockWriter{}
	a := tag.NewApplier(w, false)
	res := a.Apply("secret/app", tag.Tags{"old": "val"}, tag.Tags{"env": "prod"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Applied != 1 {
		t.Errorf("expected Applied=1, got %d", res.Applied)
	}
	if !w.called {
		t.Error("expected writer to be called")
	}
}

func TestApply_WriteError(t *testing.T) {
	w := &mockWriter{err: errors.New("vault unavailable")}
	a := tag.NewApplier(w, false)
	res := a.Apply("secret/app", tag.Tags{}, tag.Tags{"env": "prod"})
	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApply_NoChanges(t *testing.T) {
	w := &mockWriter{}
	a := tag.NewApplier(w, false)
	res := a.Apply("secret/app", tag.Tags{"env": "prod"}, tag.Tags{"env": "prod"})
	if res.Applied != 0 {
		t.Errorf("expected Applied=0, got %d", res.Applied)
	}
	if w.called {
		t.Error("writer should not be called when no changes")
	}
}
