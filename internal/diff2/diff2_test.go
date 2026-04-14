package diff2

import (
	"strings"
	"testing"
)

func TestCompare_NoChanges(t *testing.T) {
	src := map[string]map[string]string{
		"secret/app": {"key": "value"},
	}
	dst := map[string]map[string]string{
		"secret/app": {"key": "value"},
	}
	r := Compare(src, dst)
	if r.HasChanges() {
		t.Fatalf("expected no changes, got %d", len(r.Deltas))
	}
}

func TestCompare_Added(t *testing.T) {
	src := map[string]map[string]string{}
	dst := map[string]map[string]string{
		"secret/app": {"token": "abc"},
	}
	r := Compare(src, dst)
	if len(r.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(r.Deltas))
	}
	if r.Deltas[0].Type != Added {
		t.Errorf("expected Added, got %s", r.Deltas[0].Type)
	}
	if r.Deltas[0].NewValue != "abc" {
		t.Errorf("unexpected new value: %s", r.Deltas[0].NewValue)
	}
}

func TestCompare_Removed(t *testing.T) {
	src := map[string]map[string]string{
		"secret/app": {"token": "abc"},
	}
	dst := map[string]map[string]string{}
	r := Compare(src, dst)
	if len(r.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(r.Deltas))
	}
	if r.Deltas[0].Type != Removed {
		t.Errorf("expected Removed, got %s", r.Deltas[0].Type)
	}
}

func TestCompare_Changed(t *testing.T) {
	src := map[string]map[string]string{
		"secret/db": {"pass": "old"},
	}
	dst := map[string]map[string]string{
		"secret/db": {"pass": "new"},
	}
	r := Compare(src, dst)
	if len(r.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(r.Deltas))
	}
	d := r.Deltas[0]
	if d.Type != Changed {
		t.Errorf("expected Changed, got %s", d.Type)
	}
	if d.OldValue != "old" || d.NewValue != "new" {
		t.Errorf("unexpected values: old=%s new=%s", d.OldValue, d.NewValue)
	}
}

func TestFormat_NoChanges(t *testing.T) {
	r := Result{}
	out := Format(r)
	if out != "no differences found" {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestFormat_ContainsSymbols(t *testing.T) {
	r := Result{
		Deltas: []Delta{
			{Path: "secret/app", Key: "token", Type: Added, NewValue: "xyz"},
			{Path: "secret/app", Key: "old", Type: Removed, OldValue: "gone"},
			{Path: "secret/db", Key: "pass", Type: Changed, OldValue: "a", NewValue: "b"},
		},
	}
	out := Format(r)
	if !strings.Contains(out, "+ secret/app#token") {
		t.Errorf("missing added line: %s", out)
	}
	if !strings.Contains(out, "- secret/app#old") {
		t.Errorf("missing removed line: %s", out)
	}
	if !strings.Contains(out, "~ secret/db#pass") {
		t.Errorf("missing changed line: %s", out)
	}
}
