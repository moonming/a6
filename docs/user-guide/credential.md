# Credential Management

The `a6 credential` command manages APISIX consumer credentials. Credentials are nested under a consumer, so all credential operations require specifying the owner consumer with `--consumer`.

## Consumer Requirement

All credential subcommands require:

| Flag | Required | Description |
|------|----------|-------------|
| `--consumer` | Yes | Consumer username that owns the credential |

Credential API paths follow this form:

```text
/apisix/admin/consumers/:username/credentials
```

## Commands

### `a6 credential list`

Lists credentials for one consumer.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--consumer` | | | Consumer username (required) |
| `--page` | | `1` | Page number |
| `--page-size` | | `20` | Page size |
| `--output` | `-o` | `table` | Output format (`table`, `json`, `yaml`) |

Examples:

```bash
a6 credential list --consumer jack
a6 credential list --consumer jack --page 2 --page-size 10
a6 credential list --consumer jack -o json
```

### `a6 credential get`

Gets one credential by ID for a specific consumer.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--consumer` | | | Consumer username (required) |
| `--output` | `-o` | `yaml` | Output format (`json`, `yaml`) |

Examples:

```bash
a6 credential get cred-1 --consumer jack
a6 credential get cred-1 --consumer jack -o json
```

### `a6 credential create`

Creates a credential from a JSON/YAML file. Create uses `PUT` with the credential `id` extracted from the file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--consumer` | | | Consumer username (required) |
| `--file` | `-f` | | Path to credential config file (required) |
| `--output` | `-o` | `yaml` | Output format (`json`, `yaml`) |

Examples:

```bash
a6 credential create --consumer jack -f credential.json
a6 credential create --consumer jack -f credential.yaml -o json
```

### `a6 credential update`

Updates an existing credential by ID using a JSON/YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--consumer` | | | Consumer username (required) |
| `--file` | `-f` | | Path to credential config file (required) |
| `--output` | `-o` | `yaml` | Output format (`json`, `yaml`) |

Examples:

```bash
a6 credential update cred-1 --consumer jack -f credential-update.json
a6 credential update cred-1 --consumer jack -f credential-update.yaml -o json
```

### `a6 credential delete`

Deletes a credential by ID for a consumer.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--consumer` | | | Consumer username (required) |
| `--force` | | `false` | Skip confirmation prompt |

Examples:

```bash
a6 credential delete cred-1 --consumer jack
a6 credential delete cred-1 --consumer jack --force
```

## Example Credential Files

Create credential (`credential.json`):

```json
{
  "id": "cred-1",
  "plugins": {
    "key-auth": {
      "key": "test-credential-key-12345"
    }
  }
}
```

Update credential (`credential-update.json`):

```json
{
  "plugins": {
    "key-auth": {
      "key": "test-credential-key-updated"
    }
  }
}
```

## Relationship to Consumers and Auth Plugins

Credentials are attached to a consumer and hold plugin-specific authentication data (for example, keys for `key-auth` or credentials for other auth plugins). A credential cannot exist independently; it must belong to an existing consumer identified by `--consumer`.
