package flatten_test

import (
	"testing"

	"github.com/your-org/vaultpatch/internal/flatten"
)

func TestFlatten_SimpleMap(t *testing.T) {
	input := map[string]any{
		"username": "admin",
		"password": "s3cr3t",
	}
	got := flatten.Flatten(input, flatten.DefaultOptions())
	if got["username"] != "admin" {
		t.Errorf("expected username=admin, got %q", got["username"])
	}
	if got["password"] != "s3cr3t" {
		t.Errorf("expected password=s3cr3t, got %q", got["password"])
	}
}

func TestFlatten_NestedMap(t *testing.T) {
	input := map[string]any{
		"db": map[string]any{
			"host": "localhost",
			"port": "5432",
		},
	}
	got := flatten.Flatten(input, flatten.DefaultOptions())
	if got["db.host"] != "localhost" {
		t.Errorf("expected db.host=localhost, got %q", got["db.host"])
	}
	if got["db.port"] != "5432" {
		t.Errorf("expected db.port=5432, got %q", got["db.port"])
	}
}

func TestFlatten_CustomSeparator(t *testing.T) {
	input := map[string]any{
		"a": map[string]any{"b": "val"},
	}
	got := flatten.Flatten(input, flatten.Options{Separator: "/"})
	if got["a/b"] != "val" {
		t.Errorf("expected a/b=val, got %v", got)
	}
}

func TestFlatten_WithPrefix(t *testing.T) {
	input := map[string]any{"key": "value"}
	got := flatten.Flatten(input, flatten.Options{Separator: ".", Prefix: "ns"})
	if got["ns.key"] != "value" {
		t.Errorf("expected ns.key=value, got %v", got)
	}
}

func TestKeys_Sorted(t *testing.T) {
	m := map[string]string{"z": "1", "a": "2", "m": "3"}
	keys := flatten.Keys(m)
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("expected sorted keys, got %v", keys)
	}
}

func TestUnflatten_Simple(t *testing.T) {
	input := map[string]string{
		"db.host": "localhost",
		"db.port": "5432",
		"app":     "myapp",
	}
	got := flatten.Unflatten(input, ".")
	db, ok := got["db"].(map[string]any)
	if !ok {
		t.Fatalf("expected db to be a map, got %T", got["db"])
	}
	if db["host"] != "localhost" {
		t.Errorf("expected db.host=localhost, got %v", db["host"])
	}
	if got["app"] != "myapp" {
		t.Errorf("expected app=myapp, got %v", got["app"])
	}
}

func TestUnflatten_EmptySeparatorDefaultsDot(t *testing.T) {
	input := map[string]string{"x.y": "z"}
	got := flatten.Unflatten(input, "")
	xMap, ok := got["x"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map under x")
	}
	if xMap["y"] != "z" {
		t.Errorf("expected x.y=z, got %v", xMap["y"])
	}
}
