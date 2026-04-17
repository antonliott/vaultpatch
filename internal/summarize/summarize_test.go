package summarize_test

import (
	"errors"
	"testing"

	"github.com/example/vaultpatch/internal/summarize"
)

type mockReader struct {
	data map[string]string
	err  error
}

func (m *mockReader) List(_ string) ([]string, error) { return nil, nil }
func (m *mockReader) Read(_ string) (map[string]string, error) {
	return m.data, m.err
}

func TestCompute_EmptySecret(t *testing.T) {
	s := summarize.New(&mockReader{data: map[string]string{}})
	sum, err := s.Compute("secret/empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.TotalKeys != 0 {
		t.Errorf("expected 0 keys, got %d", sum.TotalKeys)
	}
	if sum.MinLen != 0 {
		t.Errorf("expected MinLen 0, got %d", sum.MinLen)
	}
}

func TestCompute_BasicStats(t *testing.T) {
	s := summarize.New(&mockReader{data: map[string]string{
		"short": "hi",
		"long":  "hello world",
	}})
	sum, err := s.Compute("secret/data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.TotalKeys != 2 {
		t.Errorf("expected 2 keys, got %d", sum.TotalKeys)
	}
	if sum.MaxLen != 11 {
		t.Errorf("expected MaxLen 11, got %d", sum.MaxLen)
	}
	if sum.MinLen != 2 {
		t.Errorf("expected MinLen 2, got %d", sum.MinLen)
	}
	expectedAvg := float64(13) / 2.0
	if sum.AvgLen != expectedAvg {
		t.Errorf("expected AvgLen %.1f, got %.1f", expectedAvg, sum.AvgLen)
	}
}

func TestCompute_TypeHints(t *testing.T) {
	s := summarize.New(&mockReader{data: map[string]string{
		"enc":  "vault:v1:abc123",
		"uid":  "550e8400-e29b-41d4-a716-446655440000",
		"link": "https://example.com",
		"plain": "hello",
	}})
	sum, err := s.Compute("secret/hints")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hints := map[string]string{}
	for _, st := range sum.Stats {
		hints[st.Key] = st.TypeHint
	}
	if hints["enc"] != "encrypted" {
		t.Errorf("expected encrypted, got %s", hints["enc"])
	}
	if hints["uid"] != "uuid" {
		t.Errorf("expected uuid, got %s", hints["uid"])
	}
	if hints["link"] != "url" {
		t.Errorf("expected url, got %s", hints["link"])
	}
	if hints["plain"] != "string" {
		t.Errorf("expected string, got %s", hints["plain"])
	}
}

func TestCompute_ReadError(t *testing.T) {
	s := summarize.New(&mockReader{err: errors.New("forbidden")})
	_, err := s.Compute("secret/x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
