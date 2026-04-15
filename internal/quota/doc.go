// Package quota implements per-path secret count quotas for Vault namespaces.
//
// A Checker is configured with a set of Rules, each binding a KV path to a
// maximum number of secret keys. Calling Check lists each path and returns
// Violations for any path that exceeds its limit.
//
// Example usage:
//
//	rules := []quota.Rule{
//		{Path: "secret/prod/db", Limit: 20},
//		{Path: "secret/prod/api", Limit: 10},
//	}
//	checker := quota.NewChecker(vaultReader, rules)
//	violations, err := checker.Check()
package quota
