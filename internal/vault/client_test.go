package vault

import (
	"testing"
)

func TestNewClient_MissingAddress(t *testing.T) {
	_, err := NewClient(Config{
		Token:     "root",
		Namespace: "ns1",
	})
	if err == nil {
		t.Fatal("expected error when address is empty, got nil")
	}
}

func TestNewClient_MissingToken(t *testing.T) {
	_, err := NewClient(Config{
		Address:   "http://127.0.0.1:8200",
		Namespace: "ns1",
	})
	if err == nil {
		t.Fatal("expected error when token is empty, got nil")
	}
}

func TestNewClient_ValidConfig(t *testing.T) {
	client, err := NewClient(Config{
		Address:   "http://127.0.0.1:8200",
		Token:     "root",
		Namespace: "engineering",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Namespace() != "engineering" {
		t.Errorf("expected namespace %q, got %q", "engineering", client.Namespace())
	}
}

func TestNewClient_EmptyNamespace(t *testing.T) {
	client, err := NewClient(Config{
		Address: "http://127.0.0.1:8200",
		Token:   "root",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Namespace() != "" {
		t.Errorf("expected empty namespace, got %q", client.Namespace())
	}
}
