# Declarative Configuration

The `a6 config` command group provides tools to export and validate APISIX declarative configuration files.

## Config File Format

The declarative config file supports YAML and JSON and uses this top-level structure:

```yaml
version: "1"
routes: []
services: []
upstreams: []
consumers: []
ssl: []
global_rules: []
plugin_configs: []
consumer_groups: []
stream_routes: []
protos: []
secrets: []
plugin_metadata: []
```

Supported resource sections:

- `routes`
- `services`
- `upstreams`
- `consumers`
- `ssl`
- `global_rules`
- `plugin_configs`
- `consumer_groups`
- `stream_routes`
- `protos`
- `secrets`
- `plugin_metadata`

Notes:

- `version` must be `"1"`.
- `create_time` and `update_time` are excluded from dumped output.
- `plugin_metadata` entries include `plugin_name` and plugin-specific metadata fields.
- `secrets` IDs use compound format such as `vault/my-vault`.

## `a6 config dump`

Dump resources from APISIX Admin API into a declarative config file.

```bash
a6 config dump [--output yaml|json] [--file output.yaml]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (`yaml`, `json`) |
| `--file` | `-f` | | Write output to file instead of stdout |

Examples:

```bash
# Dump as YAML to stdout
a6 config dump

# Dump as JSON to stdout
a6 config dump -o json

# Dump to file
a6 config dump -f apisix-config.yaml
```

## `a6 config validate`

Validate a declarative config file structure.

```bash
a6 config validate -f config.yaml
```

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--file` | `-f` | Yes | Path to a YAML/JSON declarative config file |

Validation checks include:

- `version` exists and equals `"1"`
- Required fields (for example: route requires `uri` or `uris`, consumer requires `username`)
- ID format validation (alphanumeric, `.`, `_`, `-`, max 64 chars)
- Duplicate ID detection within each resource type

Examples:

```bash
# Validate a YAML file
a6 config validate -f apisix-config.yaml

# Validate a JSON file
a6 config validate -f apisix-config.json
```

On success, the command prints:

```text
Config is valid
```

## `a6 config diff`

Compare a local declarative config file with the current APISIX Admin state.

```bash
a6 config diff -f config.yaml [--output json]
```

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--file` | `-f` | Yes | | Path to local YAML/JSON declarative config file |
| `--output` | `-o` | No | | Output format (`json` for machine-readable output; empty for human-readable summary) |

Behavior:

- Compares all declarative resource sections (`routes`, `services`, `upstreams`, `consumers`, `ssl`, `global_rules`, `plugin_configs`, `consumer_groups`, `stream_routes`, `protos`, `secrets`, `plugin_metadata`)
- Classifies resources as `CREATE`, `UPDATE`, `DELETE`, or unchanged
- Uses `id` for most resources, `username` for consumers, `plugin_name` for plugin metadata

Exit codes:

- `0`: no differences
- `1`: differences found

Examples:

```bash
# Human-readable summary
a6 config diff -f apisix-config.yaml

# JSON output
a6 config diff -f apisix-config.yaml -o json
```

## `a6 config sync`

Synchronize APISIX Admin resources to match a local declarative config file.

```bash
a6 config sync -f config.yaml [--dry-run] [--delete=true|false]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--file`, `-f` | Yes | | Path to local YAML/JSON declarative config file |
| `--dry-run` | No | `false` | Show what would change without applying any API mutations |
| `--delete` | No | `true` | Delete remote resources that are not present in the local config |

Behavior:

- Validates the local config first (same checks as `a6 config validate`)
- Computes diff between local and remote state
- Applies changes in order: create, update, delete
- Uses explicit PUT with resource keys from config for create/update operations
- Prints a per-resource summary of created/updated/deleted counts

`--dry-run` behavior:

- Does not send create/update/delete mutations
- Output is the same diff summary format as `a6 config diff`

Examples:

```bash
# Apply full sync (create/update/delete)
a6 config sync -f apisix-config.yaml

# Preview changes only
a6 config sync -f apisix-config.yaml --dry-run

# Create/update only (do not delete extra remote resources)
a6 config sync -f apisix-config.yaml --delete=false
```
