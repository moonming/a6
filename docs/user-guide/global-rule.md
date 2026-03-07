# Global Rule Management

The `a6 global-rule` command allows you to manage Apache APISIX global rules. You can list, create, update, get, and delete global rules using the CLI.

## Commands

### `a6 global-rule list`

Lists all global rules in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all global rules:
```bash
a6 global-rule list
```

Output in JSON format:
```bash
a6 global-rule list -o json
```

Paginated output:
```bash
a6 global-rule list --page 2 --page-size 5
```

### `a6 global-rule get`

Gets detailed information about a specific global rule by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get global rule by ID:
```bash
a6 global-rule get 1
```

Get global rule in JSON format:
```bash
a6 global-rule get 1 -o json
```

### `a6 global-rule create`

Creates a new global rule from a JSON or YAML file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the global rule configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create a global rule from a JSON file:
```bash
a6 global-rule create -f global-rule.json
```

Create a global rule from a YAML file:
```bash
a6 global-rule create -f global-rule.yaml
```

### `a6 global-rule update`

Updates an existing global rule using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the global rule configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update global rule with ID `1`:
```bash
a6 global-rule update 1 -f updated-global-rule.json
```

### `a6 global-rule delete`

Deletes a global rule by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete global rule with confirmation:
```bash
a6 global-rule delete 1
```

Delete global rule without confirmation:
```bash
a6 global-rule delete 1 --force
```

## Global Rule Configuration Reference

Key fields in the global rule configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the global rule |
| `plugins` | Plugin configurations for the global rule |
| `create_time` | Resource creation timestamp (Unix seconds) |
| `update_time` | Resource last update timestamp (Unix seconds) |

## Sample Configuration

```json
{
  "id": "test-global-rule-crud-1",
  "plugins": {
    "prometheus": {}
  }
}
```

## APISIX 3.15.0 Note

In APISIX 3.15.0, the same plugin cannot be used in multiple global rules at the same time. Use unique plugin sets per global rule to avoid Admin API validation errors.
