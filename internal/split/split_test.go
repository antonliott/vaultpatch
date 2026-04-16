package split_test

import (
	"errors"
	"testing"

	"github.com/youorg/vaultpatch/internal/split"
)

type mockRW struct {
	data    map[string]map[string]string
	written map[string]map[string]string
	readErr error
}

func (m *mockRW) List(path string) ([]string, error) { return nil, nil }
func (m *mockRW) Read(path string) (map[string]string, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if v, ok := m.data[path]; ok {
		return v, nil
	}
	return map[string]string{}, nil
}
func (m *mockRW) Write(path string, data map[string]string) error {
	if m.written == nil {
		m.written = map[string]map[string]string{}
	}
	m.written[path] = data
	return nil
}

func TestApply_DryRun(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/svc": {"env": "prod", "api_key": "abc"},
	}}
	s := split.NewSplitter(rw, "env", "secret/by-env", true)
	res, err := s.Apply("secret/svc")
	if err != nil {
		t.Fatal(err)
	}
	if !res.DryRun {
		t.Error("expected dry run")
	}
	if len(res.Written) != 1 || res.Written[0] != "secret/by-env/prod" {
		t.Errorf("unexpected written: %v", res.Written)
	}
	if rw.written != nil {
		t.Error("dry run should not write")
	}
}

func TestApply_Success(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/svc": {"env": "staging", "token": "xyz"},
	}}
	s := split.NewSplitter(rw, "env", "secret/by-env", false)
	res, err := s.Apply("secret/svc")
	if err != nil {
		t.Fatal(err)
	}
	if res.DryRun {
		t.Error("expected live run")
	}
	written := rw.written["secret/by-env/staging"]
	if written["token"] != "xyz" {
		t.Errorf("expected token=xyz, got %v", written)
	}
	if _, ok := written["env"]; ok {
		t.Error("key field should be excluded from payload")
	}
}

func TestApply_MissingKeyField_Skipped(t *testing.T) {
	rw := &mockRW{data: map[string]map[string]string{
		"secret/svc": {"token": "abc"},
	}}
	s := split.NewSplitter(rw, "env", "secret/by-env", false)
	res, err := s.Apply("secret/svc")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected skip, got %+v", res)
	}
}

func TestApply_ReadError(t *testing.T) {
	rw := &mockRW{readErr: errors.New("vault down")}
	s := split.NewSplitter(rw, "env", "secret/by-env", false)
	_, err := s.Apply("secret/svc")
	if err == nil {
		t.Error("expected error")
	}
}
