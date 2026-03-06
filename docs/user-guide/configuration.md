# Configuration

The a6 CLI manages connections to various Apache APISIX instances through a flexible configuration system that supports multiple contexts, environment variables, and command line overrides.

## Config File Location

The a6 CLI stores its configuration in a YAML file. It searches for this file in several locations, following this precedence order:

1. `A6_CONFIG_DIR`: If this environment variable is set, a6 looks for `config.yaml` inside that directory.
2. `XDG_CONFIG_HOME`: If set, a6 uses `$XDG_CONFIG_HOME/a6/config.yaml`.
3. Default: `~/.config/a6/config.yaml` on most systems.

The configuration file is created lazily. It won't exist until you run your first `a6 context create` command.

## Config File Format

The `config.yaml` file stores multiple connection profiles and tracks which one is currently active.

```yaml
current-context: local
contexts:
  - name: local
    server: http://localhost:9180
    api-key: edd1c9f034335f136f87ad84b625c8f1
  - name: staging
    server: http://staging.example.com:9180
    api-key: staging-api-key-here
```

The `api-key` field is optional and may be omitted if the server doesn't require authentication.

## Environment Variables

You can control a6 behavior using environment variables. These are useful for CI/CD pipelines or temporary overrides.

| Variable | Description |
|----------|-------------|
| `A6_SERVER` | The URL of the APISIX Admin API |
| `A6_API_KEY` | The API key for authentication |
| `A6_CONFIG_DIR` | Custom directory for the config file |
| `NO_COLOR` | Disable colored output if set |

## Override Precedence

When multiple configuration sources exist, a6 determines the final value using this priority:

1. Command line flags (e.g., `--server`, `--api-key`)
2. Environment variables
3. Current context in the config file

For example, if you have a `local` context active but want to run a one-off command against a production server:

```bash
a6 route list --server http://prod.example.com:9180 --api-key prod-key
```

## Context Management

Contexts allow you to switch between different APISIX environments quickly.

### create

Create a new context profile.

**Usage:** `a6 context create <name> --server <url> [--api-key <key>]`

The first context you create is automatically set as the current context.

```bash
a6 context create prod --server http://1.2.3.4:9180 --api-key secret
```

### use

Switch to a different context.

**Usage:** `a6 context use <name>`

```bash
a6 context use staging
# Output: switched to context "staging"
```

### list

List all available contexts.

**Usage:** `a6 context list` (alias: `ls`)

You can use the global `--output` flag to get machine-readable data.

```bash
a6 context list --output json
```

### delete

Remove a context from the configuration.

**Usage:** `a6 context delete <name> [--force]` (alias: `rm`)

If you delete the current context, a6 automatically switches to the first remaining context. Non-TTY environments skip the confirmation prompt.

```bash
a6 context delete staging --force
# Output: context "staging" deleted
```

### current

Display the name of the active context.

**Usage:** `a6 context current`

```bash
a6 context current
# Output: local
```

## Examples

### Single Instance Setup

For a simple local setup, create one context and start managing resources.

```bash
a6 context create local --server http://localhost:9180 --api-key edd1c9f034335f136f87ad84b625c8f1
a6 route list
```

### Multi-Environment Management

Switch between development and staging environments.

```bash
a6 context create dev --server http://dev.apisix:9180
a6 context create staging --server http://staging.apisix:9180
a6 context use staging
a6 route list
```

### CI/CD Usage

In automated environments, use environment variables to avoid creating configuration files on disk.

```bash
export A6_SERVER="http://apisix-prod:9180"
export A6_API_KEY="prod-secret-key"
a6 route list
```