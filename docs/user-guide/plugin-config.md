# Plugin Config Management

The `a6 plugin-config` command allows you to manage Apache APISIX plugin configs. Plugin configs are reusable plugin sets that can be referenced by other resources.

## Commands

### `a6 plugin-config list`

Lists all plugin configs in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all plugin configs:
```bash
a6 plugin-config list
```

Output in JSON format:
```bash
a6 plugin-config list -o json
```

Paginated output:
```bash
a6 plugin-config list --page 2 --page-size 5
```

### `a6 plugin-config get`

Gets detailed information about a specific plugin config by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get plugin config by ID:
```bash
a6 plugin-config get test-plugin-config-1
```

Get plugin config in JSON format:
```bash
a6 plugin-config get test-plugin-config-1 -o json
```

### `a6 plugin-config create`

Creates a new plugin config from a JSON or YAML file.

Creation uses `PUT /apisix/admin/plugin_configs/:id`, so the configuration file must include an `id` field.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the plugin config file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create a plugin config from JSON:
```bash
a6 plugin-config create -f plugin-config.json
```

Create a plugin config from YAML:
```bash
a6 plugin-config create -f plugin-config.yaml
```

### `a6 plugin-config update`

Updates an existing plugin config by ID using a JSON or YAML file.

Update uses `PATCH /apisix/admin/plugin_configs/:id`.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the plugin config file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update plugin config with ID `test-plugin-config-1`:
```bash
a6 plugin-config update test-plugin-config-1 -f plugin-config-updated.json
```

### `a6 plugin-config delete`

Deletes a plugin config by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete plugin config with confirmation:
```bash
a6 plugin-config delete test-plugin-config-1
```

Delete plugin config without confirmation:
```bash
a6 plugin-config delete test-plugin-config-1 --force
```

## Plugin Config Configuration Reference

Key fields in plugin config:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the plugin config (required for create) |
| `name` | Human-readable name |
| `desc` | Description |
| `plugins` | Reusable plugin configuration set |
| `labels` | User-defined labels |
| `create_time` | Resource creation timestamp (Unix seconds) |
| `update_time` | Resource last update timestamp (Unix seconds) |

## Sample JSON

```json
{
  "id": "test-plugin-config-1",
  "name": "rate-limit-set",
  "desc": "Shared rate limit plugins",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  },
  "labels": {
    "env": "test"
  }
}
```

## Sample YAML

```yaml
id: test-plugin-config-1
name: rate-limit-set
desc: Shared rate limit plugins
plugins:
  limit-count:
    count: 100
    time_window: 60
    rejected_code: 503
    key_type: var
    key: remote_addr
labels:
  env: test
```
