package group_test

import (
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/group"
)

type mockReader struct {
	paths   []string
	secrets map[string]map[string]string
	listErr error
	readErr error
}

func (m *mockReader) List(_ string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.paths, nil
}

func (m *mockReader) Read(path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.secrets[path], nil
}

func TestApply_NoMatch(t *testing.T) {
	r := &mockReader{
		paths: []string{"app"},
		secrets: map[string]map[string]string{
			"secret/app": {"db_host": "localhost"},
		},
	}
	g := group.NewGrouper(r)
	groups, err := g.Apply("secret", "aws_")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestApply_Success(t *testing.T) {
	r := &mockReader{
		paths: []string{"app", "worker"},
		secrets: map[string]map[string]string{
			"secret/app":    {"db_host": "localhost", "db_pass": "s3cr3t"},
			"secret/worker": {"db_pass": "other", "queue_url": "sqs://"},
		},
	}
	g := group.NewGrouper(r)
	groups, err := g.Apply("secret", "db_")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Prefix != "db_" {
		t.Errorf("unexpected prefix: %s", groups[0].Prefix)
	}
	if len(groups[0].Entries) != 2 {
		t.Errorf("expected 2 paths in group, got %d", len(groups[0].Entries))
	}
}

func TestApply_ListError(t *testing.T) {
	r := &mockReader{listErr: errors.New("permission denied")}
	g := group.NewGrouper(r)
	_, err := g.Apply("secret", "db_")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApply_ReadError(t *testing.T) {
	r := &mockReader{
		paths:   []string{"app"},
		readErr: errors.New("read failed"),
	}
	g := group.NewGrouper(r)
	_, err := g.Apply("secret", "db_")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormat_NoGroups(t *testing.T) {
	out := group.Format(nil)
	if out != "no matching keys found\n" {
		t.Errorf("unexpected output: %q", out)
	}
}
