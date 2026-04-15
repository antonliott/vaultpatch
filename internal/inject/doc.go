// Package inject resolves vault:// references embedded in key-value maps,
// fetching the referenced secret fields from Vault and substituting them
// in-place. This is useful for injecting Vault secrets into environment
// variable maps before writing them to a .env file or passing them to a
// subprocess.
//
// Reference syntax:
//
//	vault://<mount/path>#<field>
//
// Example:
//
//	DB_PASSWORD=vault://secret/myapp/db#password
package inject
