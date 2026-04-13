package search_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/search"
)

type mockReader struct {
	paths   []string
	secrets map[string]map[string]string
	listErr error
}

func (m *mockReader) List(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.paths, nil
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]string, error) {
	if s, ok := m.secrets[path]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func TestSearch_NoFilter(t *testing.T) {
	r := &mockReader{
		paths: []string{"secret/app"},
		secrets: map[string]map[string]string{
			"secret/app": {"db_pass": "hunter2", "api_key": "abc123"},
		},
	}
	s, err := search.NewSearcher(r, "secret", "", "")
	if err != nil {
		t.Fatal(err)
	}
	matches, err := s.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestSearch_KeyFilter(t *testing.T) {
	r := &mockReader{
		paths: []string{"secret/app"},
		secrets: map[string]map[string]string{
			"secret/app": {"db_pass": "hunter2", "api_key": "abc123"},
		},
	}
	s, err := search.NewSearcher(r, "secret", "db_", "")
	if err != nil {
		t.Fatal(err)
	}
	matches, _ := s.Run(context.Background())
	if len(matches) != 1 || matches[0].Key != "db_pass" {
		t.Fatalf("unexpected matches: %+v", matches)
	}
}

func TestSearch_ValueFilter(t *testing.T) {
	r := &mockReader{
		paths: []string{"secret/app"},
		secrets: map[string]map[string]string{
			"secret/app": {"db_pass": "hunter2", "api_key": "abc123"},
		},
	}
	s, err := search.NewSearcher(r, "secret", "", "^abc")
	if err != nil {
		t.Fatal(err)
	}
	matches, _ := s.Run(context.Background())
	if len(matches) != 1 || matches[0].Key != "api_key" {
		t.Fatalf("unexpected matches: %+v", matches)
	}
}

func TestSearch_ListError(t *testing.T) {
	r := &mockReader{listErr: errors.New("permission denied")}
	s, _ := search.NewSearcher(r, "secret", "", "")
	_, err := s.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearch_InvalidPattern(t *testing.T) {
	r := &mockReader{}
	_, err := search.NewSearcher(r, "secret", "[", "")
	if err == nil {
		t.Fatal("expected regexp compile error")
	}
}

func TestFormat_MaskValues(t *testing.T) {
	matches := []search.Match{{Path: "secret/app", Key: "db_pass", Value: "hunter2"}}
	var buf bytes.Buffer
	search.Format(&buf, matches, true)
	if bytes.Contains(buf.Bytes(), []byte("hunter2")) {
		t.Fatal("value should be masked")
	}
}

func TestFormat_NoMatches(t *testing.T) {
	var buf bytes.Buffer
	search.Format(&buf, nil, false)
	if !bytes.Contains(buf.Bytes(), []byte("no matches")) {
		t.Fatal("expected no matches message")
	}
}
