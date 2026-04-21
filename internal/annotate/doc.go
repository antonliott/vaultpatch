// Package annotate provides an Annotator that can add, update, and remove
// key-value annotations on existing Vault secret paths.
//
// Annotations are stored as regular secret fields alongside existing data.
// The Annotator supports dry-run mode, which computes changes without
// persisting them to Vault.
//
// Example usage:
//
//	a := annotate.NewAnnotator(vaultWriter)
//	res := a.Apply(ctx, annotate.Options{
//		Path:        "secret/myapp",
//		Annotations: map[string]string{"env": "prod", "owner": "team-a"},
//		RemoveKeys:  []string{"deprecated"},
//	})
package annotate
