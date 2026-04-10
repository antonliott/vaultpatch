package vault

import (
	"errors"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the HashiCorp Vault API client with namespace support.
type Client struct {
	api       *vaultapi.Client
	namespace string
}

// Config holds the configuration required to create a Vault client.
type Config struct {
	Address   string
	Token     string
	Namespace string
}

// NewClient creates and returns a new authenticated Vault client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("vault address must not be empty")
	}
	if cfg.Token == "" {
		return nil, errors.New("vault token must not be empty")
	}

	apiCfg := vaultapi.DefaultConfig()
	apiCfg.Address = cfg.Address

	api, err := vaultapi.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		api.SetNamespace(cfg.Namespace)
	}

	return &Client{
		api:       api,
		namespace: cfg.Namespace,
	}, nil
}

// Namespace returns the namespace the client is configured for.
func (c *Client) Namespace() string {
	return c.namespace
}

// ReadSecret reads a KV v2 secret at the given path and returns its data.
func (c *Client) ReadSecret(path string) (map[string]interface{}, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at %q", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		// Fallback for KV v1
		return secret.Data, nil
	}

	kvData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at %q", path)
	}

	return kvData, nil
}

// WriteSecret writes data to a KV v2 secret at the given path.
func (c *Client) WriteSecret(path string, data map[string]interface{}) error {
	payload := map[string]interface{}{"data": data}
	_, err := c.api.Logical().Write(path, payload)
	if err != nil {
		return fmt.Errorf("writing secret at %q: %w", path, err)
	}
	return nil
}
