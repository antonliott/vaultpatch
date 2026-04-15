package normalize

import (
	"testing"
)

func TestNew_ValidStyles(t *testing.T) {
	for _, s := range []Style{StyleUpper, StyleLower, StyleSnake, StyleKebab} {
		n, err := New(s)
		if err != nil {
			t.Fatalf("New(%q) unexpected error: %v", s, err)
		}
		if n == nil {
			t.Fatalf("New(%q) returned nil", s)
		}
	}
}

func TestNew_InvalidStyle(t *testing.T) {
	_, err := New("pascal")
	if err == nil {
		t.Fatal("expected error for unknown style, got nil")
	}
}

func TestApply_Upper(t *testing.T) {
	n, _ := New(StyleUpper)
	got := n.Apply(map[string]string{"db_host": "localhost", "Api-Key": "secret"})
	assertKey(t, got, "DB_HOST", "localhost")
	assertKey(t, got, "API-KEY", "secret")
}

func TestApply_Lower(t *testing.T) {
	n, _ := New(StyleLower)
	got := n.Apply(map[string]string{"DB_HOST": "localhost", "API_KEY": "secret"})
	assertKey(t, got, "db_host", "localhost")
	assertKey(t, got, "api_key", "secret")
}

func TestApply_Snake(t *testing.T) {
	n, _ := New(StyleSnake)
	got := n.Apply(map[string]string{"db-host": "localhost", "API KEY": "secret"})
	assertKey(t, got, "db_host", "localhost")
	assertKey(t, got, "api_key", "secret")
}

func TestApply_Kebab(t *testing.T) {
	n, _ := New(StyleKebab)
	got := n.Apply(map[string]string{"db_host": "localhost", "API KEY": "secret"})
	assertKey(t, got, "db-host", "localhost")
	assertKey(t, got, "api-key", "secret")
}

func TestApply_DoesNotMutateInput(t *testing.T) {
	input := map[string]string{"MY_KEY": "val"}
	n, _ := New(StyleLower)
	n.Apply(input)
	if _, ok := input["my_key"]; ok {
		t.Fatal("Apply mutated the input map")
	}
	if input["MY_KEY"] != "val" {
		t.Fatal("Apply modified original key value")
	}
}

func TestApply_TrimsWhitespace(t *testing.T) {
	n, _ := New(StyleUpper)
	got := n.Apply(map[string]string{"  my_key  ": "value"})
	assertKey(t, got, "MY_KEY", "value")
}

func assertKey(t *testing.T, m map[string]string, key, want string) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("key %q not found in map %v", key, m)
		return
	}
	if v != want {
		t.Errorf("key %q: got %q, want %q", key, v, want)
	}
}
