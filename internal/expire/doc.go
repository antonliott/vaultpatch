// Package expire scans HashiCorp Vault KV secrets for expiry metadata and
// reports secrets that have already expired or will expire within a configurable
// warning window.
//
// Usage:
//
//	checker := expire.NewChecker(vaultReader, "expires_at", 7*24*time.Hour)
//	findings, err := checker.Scan("secret/myapp")
//	fmt.Println(expire.Format(findings))
package expire
