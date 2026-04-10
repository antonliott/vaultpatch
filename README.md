# vaultpatch

> CLI tool to diff and apply secrets changes across HashiCorp Vault namespaces

---

## Installation

```bash
go install github.com/yourorg/vaultpatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/vaultpatch.git
cd vaultpatch
go build -o vaultpatch .
```

---

## Usage

**Diff secrets between two namespaces:**

```bash
vaultpatch diff --src namespace/prod --dst namespace/staging
```

**Apply changes from a diff to a target namespace:**

```bash
vaultpatch apply --src namespace/prod --dst namespace/staging --dry-run
vaultpatch apply --src namespace/prod --dst namespace/staging
```

**Common flags:**

| Flag | Description |
|------|-------------|
| `--addr` | Vault server address (default: `$VAULT_ADDR`) |
| `--token` | Vault token (default: `$VAULT_TOKEN`) |
| `--dry-run` | Preview changes without applying them |
| `--output` | Output format: `text`, `json`, `yaml` |

**Example output:**

```
~ secret/db/password   [changed]
+ secret/api/key       [added]
- secret/old/token     [removed]
```

---

## Requirements

- Go 1.21+
- HashiCorp Vault 1.10+
- A valid Vault token with read/write permissions on target namespaces

---

## License

[MIT](LICENSE) © yourorg