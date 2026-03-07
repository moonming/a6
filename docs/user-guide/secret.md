# Secret Manager Management

The `a6 secret` command allows you to manage Apache APISIX secret manager resources. These resources store external secret backend configurations used by APISIX.

Supported manager types:

- `vault`
- `aws`
- `gcp`

Secret manager IDs use a compound format: `<manager>/<id>`. For example:

- `vault/my-vault-1`
- `aws/my-aws-1`
- `gcp/my-gcp-1`

## Commands

### `a6 secret list`

Lists all secret managers across all manager types in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all secrets:
```bash
a6 secret list
```

List with JSON output:
```bash
a6 secret list -o json
```

List page 2 with page size 10:
```bash
a6 secret list --page 2 --page-size 10
```

### `a6 secret get`

Gets a secret manager configuration by compound ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get a Vault secret manager:
```bash
a6 secret get vault/my-vault-1
```

Get in JSON output:
```bash
a6 secret get vault/my-vault-1 -o json
```

### `a6 secret create`

Creates a secret manager using a compound ID and a JSON/YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the secret configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create a Vault secret manager:
```bash
a6 secret create vault/my-vault-1 -f vault-secret.json
```

Create from YAML:
```bash
a6 secret create vault/my-vault-1 -f vault-secret.yaml
```

### `a6 secret update`

Updates an existing secret manager using a compound ID and a JSON/YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the secret configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update a Vault secret manager:
```bash
a6 secret update vault/my-vault-1 -f vault-secret-update.json
```

### `a6 secret delete`

Deletes a secret manager by compound ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete with confirmation:
```bash
a6 secret delete vault/my-vault-1
```

Delete without confirmation:
```bash
a6 secret delete vault/my-vault-1 --force
```

## Sample Configuration

Vault secret manager example (`vault-secret.json`):

```json
{
  "uri": "http://127.0.0.1:8200",
  "prefix": "/apisix/kv",
  "token": "test-token-12345"
}
```

AWS secret manager example (`aws-secret.json`):

```json
{
  "region": "us-east-1",
  "access_key_id": "AKIAEXAMPLE",
  "secret_access_key": "secret-example-value",
  "endpoint_url": "https://secretsmanager.us-east-1.amazonaws.com"
}
```
