// Package health provides Vault connectivity and status checks.
package health

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Status holds the result of a Vault health check.
type Status struct {
	Address     string
	Namespace   string
	Initialized bool
	Sealed      bool
	Version     string
	CheckedAt   time.Time
}

// String returns a human-readable summary of the health status.
func (s Status) String() string {
	state := "unsealed"
	if s.Sealed {
		state = "sealed"
	}
	init := "initialized"
	if !s.Initialized {
		init = "uninitialized"
	}
	return fmt.Sprintf("vault %s at %s (namespace: %q) — %s, %s",
		s.Version, s.Address, s.Namespace, init, state)
}

// Checker performs health checks against a Vault client.
type Checker struct {
	client    *vaultapi.Client
	namespace string
}

// NewChecker creates a Checker from the provided Vault API client.
func NewChecker(client *vaultapi.Client, namespace string) *Checker {
	return &Checker{client: client, namespace: namespace}
}

// Check queries the Vault health endpoint and returns a Status.
func (c *Checker) Check(ctx context.Context) (*Status, error) {
	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	return &Status{
		Address:     c.client.Address(),
		Namespace:   c.namespace,
		Initialized: health.Initialized,
		Sealed:      health.Sealed,
		Version:     health.Version,
		CheckedAt:   time.Now().UTC(),
	}, nil
}
