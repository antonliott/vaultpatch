// Package health provides Vault cluster health checking.
package health

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// StatusType represents the health state of a Vault node.
type StatusType int

const (
	StatusUnsealed      StatusType = iota
	StatusSealed
	StatusUninitialized
)

func (s StatusType) String() string {
	switch s {
	case StatusUnsealed:
		return "unsealed"
	case StatusSealed:
		return "sealed"
	case StatusUninitialized:
		return "uninitialized"
	default:
		return "unknown"
	}
}

// Status holds the result of a health check against a Vault node.
type Status struct {
	Address     string
	State       StatusType
	Version     string
	ClusterName string
	CheckedAt   time.Time
}

// Checker performs health checks against a Vault client.
type Checker struct {
	client *vaultapi.Client
}

// NewChecker creates a new Checker using the provided Vault API client.
func NewChecker(client *vaultapi.Client) *Checker {
	return &Checker{client: client}
}

// Check queries the Vault health endpoint and returns a Status.
func (c *Checker) Check(ctx context.Context) (*Status, error) {
	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	state := StatusUnsealed
	if !health.Initialized {
		state = StatusUninitialized
	} else if health.Sealed {
		state = StatusSealed
	}

	return &Status{
		Address:     c.client.Address(),
		State:       state,
		Version:     health.Version,
		ClusterName: health.ClusterName,
		CheckedAt:   time.Now().UTC(),
	}, nil
}
