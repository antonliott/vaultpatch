// Package tag provides functionality for managing metadata tags on
// HashiCorp Vault KV secrets.
//
// Tags are arbitrary string key/value pairs stored in the Vault KV v2
// metadata endpoint. This package supports comparing current tags against
// a desired state, computing deltas, and applying changes with optional
// dry-run support.
//
// Typical usage:
//
//	applier := tag.NewApplier(vaultWriter, dryRun)
//	result := applier.Apply(path, currentTags, desiredTags)
package tag
