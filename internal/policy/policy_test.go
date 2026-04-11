package policy

import (
	"strings"
	"testing"
)

func TestCompare_NoChanges(t *testing.T) {
	src := map[string]string{"admin": `path "secret/*" { capabilities = ["read"] }`}
	tgt := map[string]string{"admin": `path "secret/*" { capabilities = ["read"] }`}

	deltas := Compare(src, tgt)
	if len(deltas) != 0 {
		t.Fatalf("expected 0 deltas, got %d", len(deltas))
	}
}

func TestCompare_Added(t *testing.T) {
	src := map[string]string{"new-policy": `path "kv/*" { capabilities = ["create"] }`}
	tgt := map[string]string{}

	deltas := Compare(src, tgt)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Type != DiffAdded {
		t.Errorf("expected DiffAdded, got %s", deltas[0].Type)
	}
	if deltas[0].Name != "new-policy" {
		t.Errorf("unexpected name: %s", deltas[0].Name)
	}
}

func TestCompare_Removed(t *testing.T) {
	src := map[string]string{}
	tgt := map[string]string{"old-policy": `path "kv/*" { capabilities = ["delete"] }`}

	deltas := Compare(src, tgt)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Type != DiffRemoved {
		t.Errorf("expected DiffRemoved, got %s", deltas[0].Type)
	}
}

func TestCompare_Changed(t *testing.T) {
	src := map[string]string{"admin": `path "secret/*" { capabilities = ["read", "write"] }`}
	tgt := map[string]string{"admin": `path "secret/*" { capabilities = ["read"] }`}

	deltas := Compare(src, tgt)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Type != DiffChanged {
		t.Errorf("expected DiffChanged, got %s", deltas[0].Type)
	}
}

func TestFormat_NoDeltas(t *testing.T) {
	out := Format(nil)
	if out != "no policy differences found" {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestFormat_WithDeltas(t *testing.T) {
	deltas := []Delta{
		{Name: "alpha", Type: DiffAdded},
		{Name: "beta", Type: DiffRemoved},
	}
	out := Format(deltas)
	if !strings.Contains(out, "[added] alpha") {
		t.Errorf("missing added line in output: %s", out)
	}
	if !strings.Contains(out, "[removed] beta") {
		t.Errorf("missing removed line in output: %s", out)
	}
}
