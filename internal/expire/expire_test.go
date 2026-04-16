package expire_test

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultpatch/internal/expire"
)

type mockReader struct {
	keys map[string][]string
	data map[string]map[string]string
}

func (m *mockReader) List(path string) ([]string, error) {
	v, ok := m.keys[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (m *mockReader) Read(path string) (map[string]string, error) {
	v, ok := m.data[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func TestScan_NoExpiryKeys(t *testing.T) {
	r := &mockReader{
		keys: map[string][]string{"secret/app": {"db"}},
		data: map[string]map[string]string{"secret/app/db": {"password": "s3cr3t"}},
	}
	c := expire.NewChecker(r, "expires_at", 24*time.Hour)
	findings, err := c.Scan("secret/app")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestScan_DetectsExpired(t *testing.T) {
	past := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	r := &mockReader{
		keys: map[string][]string{"secret/app": {"token"}},
		data: map[string]map[string]string{"secret/app/token": {"expires_at": past}},
	}
	c := expire.NewChecker(r, "", 24*time.Hour)
	findings, err := c.Scan("secret/app")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if !findings[0].Expired {
		t.Error("expected finding to be marked Expired")
	}
}

func TestScan_DetectsExpiringSoon(t *testing.T) {
	soon := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	r := &mockReader{
		keys: map[string][]string{"secret/app": {"cert"}},
		data: map[string]map[string]string{"secret/app/cert": {"expires_at": soon}},
	}
	c := expire.NewChecker(r, "", 24*time.Hour)
	findings, err := c.Scan("secret/app")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Expired {
		t.Error("expected finding NOT to be marked Expired")
	}
}

func TestScan_ListError(t *testing.T) {
	r := &mockReader{keys: map[string][]string{}, data: map[string]map[string]string{}}
	c := expire.NewChecker(r, "", time.Hour)
	_, err := c.Scan("secret/missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormat_NoFindings(t *testing.T) {
	out := expire.Format(nil)
	if out != "no expiring secrets found" {
		t.Errorf("unexpected output: %q", out)
	}
}
