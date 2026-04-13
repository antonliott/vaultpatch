package access_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/access"
)

type mockReader struct {
	accessors []string
	infos     map[string]*access.TokenInfo
	listErr   error
	lookupErr error
}

func (m *mockReader) ListAccessors(_ context.Context) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.accessors, nil
}

func (m *mockReader) LookupAccessor(_ context.Context, acc string) (*access.TokenInfo, error) {
	if m.lookupErr != nil {
		return nil, m.lookupErr
	}
	info, ok := m.infos[acc]
	if !ok {
		return nil, errors.New("not found")
	}
	return info, nil
}

func TestCollect_Success(t *testing.T) {
	mr := &mockReader{
		accessors: []string{"abc123"},
		infos: map[string]*access.TokenInfo{
			"abc123": {Accessor: "abc123", DisplayName: "ci-token", TTL: 3600},
		},
	}
	a := access.NewAuditor(mr)
	infos, err := a.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) != 1 || infos[0].Accessor != "abc123" {
		t.Errorf("unexpected infos: %+v", infos)
	}
}

func TestCollect_ListError(t *testing.T) {
	mr := &mockReader{listErr: errors.New("forbidden")}
	a := access.NewAuditor(mr)
	_, err := a.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollect_LookupError(t *testing.T) {
	mr := &mockReader{
		accessors: []string{"xyz"},
		lookupErr: errors.New("lookup failed"),
	}
	a := access.NewAuditor(mr)
	_, err := a.Collect(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCompare_NoDeltas(t *testing.T) {
	src := []*access.TokenInfo{{Accessor: "a1", DisplayName: "dev", Policies: []string{"read"}, TTL: 1800}}
	dst := []*access.TokenInfo{{Accessor: "a1", DisplayName: "dev", Policies: []string{"read"}, TTL: 1800}}
	deltas := access.Compare(src, dst)
	if len(deltas) != 0 {
		t.Errorf("expected no deltas, got %+v", deltas)
	}
}

func TestCompare_TTLChanged(t *testing.T) {
	src := []*access.TokenInfo{{Accessor: "a1", DisplayName: "dev", TTL: 1800}}
	dst := []*access.TokenInfo{{Accessor: "a1", DisplayName: "dev", TTL: 3600}}
	deltas := access.Compare(src, dst)
	if len(deltas) != 1 || deltas[0].Field != "ttl" {
		t.Errorf("expected ttl delta, got %+v", deltas)
	}
}

func TestCompare_PoliciesChanged(t *testing.T) {
	src := []*access.TokenInfo{{Accessor: "a1", Policies: []string{"read"}, TTL: 900}}
	dst := []*access.TokenInfo{{Accessor: "a1", Policies: []string{"read", "write"}, TTL: 900}}
	deltas := access.Compare(src, dst)
	if len(deltas) != 1 || deltas[0].Field != "policies" {
		t.Errorf("expected policies delta, got %+v", deltas)
	}
}
