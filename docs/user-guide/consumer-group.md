# Consumer Group Management

The `a6 consumer-group` command allows you to manage Apache APISIX consumer groups. Consumer groups are reusable plugin sets that can be referenced by multiple consumers through the consumer `group_id` field.

## Commands

### `a6 consumer-group list`

Lists all consumer groups in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all consumer groups:
```bash
a6 consumer-group list
```

Output in JSON format:
```bash
a6 consumer-group list -o json
```

Paginated output:
```bash
a6 consumer-group list --page 2 --page-size 5
```

### `a6 consumer-group get`

Gets detailed information about a specific consumer group by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get consumer group by ID:
```bash
a6 consumer-group get test-consumer-group-1
```

Get consumer group in JSON format:
```bash
a6 consumer-group get test-consumer-group-1 -o json
```

### `a6 consumer-group create`

Creates a new consumer group from a JSON or YAML file.

Consumer group creation is PUT-only and requires an explicit `id` in the input file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the consumer group configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Create a consumer group from a JSON file:
```bash
a6 consumer-group create -f consumer-group.json
```

Create a consumer group from a YAML file:
```bash
a6 consumer-group create -f consumer-group.yaml
```

### `a6 consumer-group update`

Updates an existing consumer group by ID using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the consumer group configuration file (required) |
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Update consumer group with ID `test-consumer-group-1`:
```bash
a6 consumer-group update test-consumer-group-1 -f updated-consumer-group.json
```

### `a6 consumer-group delete`

Deletes a consumer group by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete consumer group with confirmation:
```bash
a6 consumer-group delete test-consumer-group-1
```

Delete consumer group without confirmation:
```bash
a6 consumer-group delete test-consumer-group-1 --force
```

## Using Consumer Groups with Consumers

To bind a consumer to a consumer group, set the consumer's `group_id` field:

```json
{
  "username": "app-client-1",
  "group_id": "test-consumer-group-1"
}
```

Then apply it with:

```bash
a6 consumer update app-client-1 -f consumer-with-group.json
```

## Consumer Group Configuration Reference

Key fields in the consumer group configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the consumer group |
| `name` | Human-readable name |
| `desc` | Description |
| `plugins` | Plugin configurations shared by consumers in the group |
| `labels` | Optional key/value labels |
| `create_time` | Resource creation timestamp (Unix seconds) |
| `update_time` | Resource last update timestamp (Unix seconds) |

## Sample Configuration

```json
{
  "id": "test-consumer-group-1",
  "name": "gold",
  "desc": "Consumer group with shared rate limiting",
  "plugins": {
    "limit-count": {
      "count": 200,
      "time_window": 60,
      "rejected_code": 503,
      "key_type": "var",
      "key": "remote_addr"
    }
  },
  "labels": {
    "tier": "gold"
  }
}
```

```yaml
id: test-consumer-group-1
name: gold
desc: Consumer group with shared rate limiting
plugins:
  limit-count:
    count: 200
    time_window: 60
    rejected_code: 503
    key_type: var
    key: remote_addr
labels:
  tier: gold
```
