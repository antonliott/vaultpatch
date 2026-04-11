package diff

import (
	"strings"
	"testing"
)

func TestCompare_NoChanges(t *testing.T) {
	src := SecretMap{"key1": "val1", "key2": "val2"}
	dst := SecretMap{"key1": "val1", "key2": "val2"}
	entries := Compare(src, dst)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestCompare_Added(t *testing.T) {
	src := SecretMap{}
	dst := SecretMap{"newkey": "newval"}
	entries := Compare(src, dst)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Type != DiffAdded || entries[0].Key != "newkey" || entries[0].NewValue != "newval" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestCompare_Removed(t *testing.T) {
	src := SecretMap{"oldkey": "oldval"}
	dst := SecretMap{}
	entries := Compare(src, dst)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Type != DiffRemoved || entries[0].Key != "oldkey" || entries[0].OldValue != "oldval" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestCompare_Changed(t *testing.T) {
	src := SecretMap{"key": "old"}
	dst := SecretMap{"key": "new"}
	entries := Compare(src, dst)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Type != DiffChanged || e.OldValue != "old" || e.NewValue != "new" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestFormat_NoDiff(t *testing.T) {
	out := Format(nil)
	if out != "No differences found." {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormat_Mixed(t *testing.T) {
	entries := []DiffEntry{
		{Key: "a", Type: DiffAdded, NewValue: "1"},
		{Key: "b", Type: DiffRemoved, OldValue: "2"},
		{Key: "c", Type: DiffChanged, OldValue: "x", NewValue: "y"},
	}
	out := Format(entries)
	if !strings.Contains(out, "+ a") {
		t.Error("expected added line for 'a'")
	}
	if !strings.Contains(out, "- b") {
		t.Error("expected removed line for 'b'")
	}
	if !strings.Contains(out, "~ c") {
		t.Error("expected changed line for 'c'")
	}
}
