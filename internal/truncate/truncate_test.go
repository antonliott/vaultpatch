package truncate_test

import (
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/truncate"
)

type mockRW struct {
	listFn  func(string) ([]string, error)
	readFn  func(string) (map[string]string, error)
	writeFn func(string, map[string]string) error
}

func (m *mockRW) List(p string) ([]string, error)              { return m.listFn(p) }
func (m *mockRW) Read(p string) (map[string]string, error)     { return m.readFn(p) }
func (m *mockRW) Write(p string, d map[string]string) error    { return m.writeFn(p, d) }

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{
		listFn: func(string) ([]string, error) { return []string{"svc"}, nil },
		readFn: func(string) (map[string]string, error) {
			return map[string]string{"token": "averylongsecretvalue"}, nil
		},
		writeFn: func(string, map[string]string) error {
			t.Fatal("write must not be called in dry-run")
			return nil
		},
	}
	tr := truncate.NewTruncator(rw, truncate.Options{MaxLen: 5, DryRun: true})
	res, err := tr.Apply("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Truncated != "avery..." {
		t.Errorf("unexpected truncated value: %s", res[0].Truncated)
	}
}

func TestApply_Success(t *testing.T) {
	written := map[string]string{}
	rw := &mockRW{
		listFn: func(string) ([]string, error) { return []string{"svc"}, nil },
		readFn: func(string) (map[string]string, error) {
			return map[string]string{"token": "toolongvalue", "ok": "hi"}, nil
		},
		writeFn: func(_ string, d map[string]string) error { written = d; return nil },
	}
	tr := truncate.NewTruncator(rw, truncate.Options{MaxLen: 4})
	res, err := tr.Apply("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if written["token"] != "tool..." {
		t.Errorf("unexpected written value: %s", written["token"])
	}
	if written["ok"] != "hi" {
		t.Errorf("short value should be unchanged")
	}
}

func TestApply_NoChanges(t *testing.T) {
	rw := &mockRW{
		listFn: func(string) ([]string, error) { return []string{"svc"}, nil },
		readFn: func(string) (map[string]string, error) {
			return map[string]string{"k": "short"}, nil
		},
		writeFn: func(string, map[string]string) error {
			t.Fatal("write must not be called when nothing changes")
			return nil
		},
	}
	tr := truncate.NewTruncator(rw, truncate.Options{MaxLen: 100})
	res, err := tr.Apply("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected no results")
	}
}

func TestApply_ListError(t *testing.T) {
	rw := &mockRW{
		listFn: func(string) ([]string, error) { return nil, errors.New("permission denied") },
	}
	tr := truncate.NewTruncator(rw, truncate.Options{MaxLen: 10})
	_, err := tr.Apply("secret")
	if err == nil {
		t.Fatal("expected error")
	}
}
