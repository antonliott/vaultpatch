package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func TestStatus_String_Unsealed(t *testing.T) {
	if got := StatusUnsealed.String(); got != "unsealed" {
		t.Errorf("expected unsealed, got %s", got)
	}
}

func TestStatus_String_Sealed(t *testing.T) {
	if got := StatusSealed.String(); got != "sealed" {
		t.Errorf("expected sealed, got %s", got)
	}
}

func TestStatus_String_Uninitialized(t *testing.T) {
	if got := StatusUninitialized.String(); got != "uninitialized" {
		t.Errorf("expected uninitialized, got %s", got)
	}
}

func newTestServer(t *testing.T, body interface{}, code int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func newVaultClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return client
}

func contains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func TestCheck_Unsealed(t *testing.T) {
	payload := map[string]interface{}{
		"initialized":  true,
		"sealed":       false,
		"version":      "1.15.0",
		"cluster_name": "vault-cluster",
	}
	srv := newTestServer(t, payload, http.StatusOK)
	defer srv.Close()

	checker := NewChecker(newVaultClient(t, srv.URL))
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.State != StatusUnsealed {
		t.Errorf("expected unsealed, got %s", status.State)
	}
	contains(t, status.Version, "1.15.0")
	contains(t, status.ClusterName, "vault-cluster")
	if status.CheckedAt.IsZero() {
		t.Error("expected non-zero CheckedAt")
	}
}
