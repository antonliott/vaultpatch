package labelset_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/labelset"
)

func TestCompare_NoChanges(t *testing.T) {
	src := labelset.Labels{"env": "prod", "team": "platform"}
	dst := labelset.Labels{"env": "prod", "team": "platform"}
	deltas := labelset.Compare(src, dst)
	if len(deltas) != 0 {
		t.Fatalf("expected no deltas, got %d", len(deltas))
	}
}

func TestCompare_Added(t *testing.T) {
	src := labelset.Labels{}
	dst := labelset.Labels{"env": "staging"}
	deltas := labelset.Compare(src, dst)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].ChangeType != "added" || deltas[0].Key != "env" || deltas[0].New != "staging" {
		t.Errorf("unexpected delta: %+v", deltas[0])
	}
}

func TestCompare_Removed(t *testing.T) {
	src := labelset.Labels{"owner": "alice"}
	dst := labelset.Labels{}
	deltas := labelset.Compare(src, dst)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].ChangeType != "removed" || deltas[0].Key != "owner" || deltas[0].Old != "alice" {
		t.Errorf("unexpected delta: %+v", deltas[0])
	}
}

func TestCompare_Changed(t *testing.T) {
	src := labelset.Labels{"env": "dev"}
	dst := labelset.Labels{"env": "prod"}
	deltas := labelset.Compare(src, dst)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	d := deltas[0]
	if d.ChangeType != "changed" || d.Old != "dev" || d.New != "prod" {
		t.Errorf("unexpected delta: %+v", d)
	}
}

func TestFormat_NoDeltas(t *testing.T) {
	out := labelset.Format(nil)
	if out != "no label changes" {
		t.Errorf("expected 'no label changes', got %q", out)
	}
}

func TestFormat_WithDeltas(t *testing.T) {
	deltas := []labelset.Delta{
		{Key: "env", New: "prod", ChangeType: "added"},
		{Key: "owner", Old: "bob", ChangeType: "removed"},
		{Key: "team", Old: "sre", New: "platform", ChangeType: "changed"},
	}
	out := labelset.Format(deltas)
	if !strings.Contains(out, "+ env") {
		t.Errorf("expected added label in output, got: %s", out)
	}
	if !strings.Contains(out, "- owner") {
		t.Errorf("expected removed label in output, got: %s", out)
	}
	if !strings.Contains(out, "~ team") {
		t.Errorf("expected changed label in output, got: %s", out)
	}
}
