package linkage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/vaultpatch/internal/linkage"
)

type mockReader struct {
	paths  []string
	secrets map[string]map[string]string
	listErr error
	readErr error
}

func (m *mockReader) List(_ context.Context, _ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.paths, nil
}

func (m *mockReader) Read(_ context.Context, path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.secrets[path], nil
}

func TestDetect_NoLinks(t *testing.T) {
	r := &mockReader{
		paths: []string{"alpha", "beta"},
		secrets: map[string]map[string]string{
			"secret/alpha": {"key": "value"},
			"secret/beta":  {"key": "other"},
		},
	}
	d := linkage.NewDetector(r, "secret")
	links, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestDetect_FindsLink(t *testing.T) {
	r := &mockReader{
		paths: []string{"alpha", "beta"},
		secrets: map[string]map[string]string{
			"secret/alpha": {"ref": "secret/beta"},
			"secret/beta":  {"key": "value"},
		},
	}
	d := linkage.NewDetector(r, "secret")
	links, err := d.Detect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].SourcePath != "secret/alpha" || links[0].SourceKey != "ref" || links[0].TargetPath != "secret/beta" {
		t.Errorf("unexpected link: %+v", links[0])
	}
}

func TestDetect_ListError(t *testing.T) {
	r := &mockReader{listErr: errors.New("permission denied")}
	d := linkage.NewDetector(r, "secret")
	_, err := d.Detect(context.Background())
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected list error, got %v", err)
	}
}

func TestDetect_ReadError(t *testing.T) {
	r := &mockReader{
		paths:   []string{"alpha"},
		readErr: errors.New("read failed"),
	}
	d := linkage.NewDetector(r, "secret")
	_, err := d.Detect(context.Background())
	if err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestFormat_NoLinks(t *testing.T) {
	out := linkage.Format(nil)
	if !strings.Contains(out, "no cross-path") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormat_WithLinks(t *testing.T) {
	links := []linkage.Link{
		{SourcePath: "secret/alpha", SourceKey: "ref", TargetPath: "secret/beta"},
	}
	out := linkage.Format(links)
	if !strings.Contains(out, "secret/alpha[ref] -> secret/beta") {
		t.Errorf("unexpected output: %q", out)
	}
}
