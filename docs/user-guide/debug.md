# Debug and Tracing

The `a6 debug` command group helps diagnose APISIX request behavior.

## `a6 debug trace`

Trace a request through a specific route by fetching route configuration from the Admin API, optionally reading plugin priorities from the Control API, then sending a real request through the APISIX gateway.

```bash
a6 debug trace <route-id>
```

When `<route-id>` is omitted in a terminal session, `a6` opens an interactive route picker. In non-interactive mode, route ID is required.

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--method` | | route first method or `GET` | HTTP method for the probe request |
| `--path` | | route `uri` | Request path for the probe request |
| `--header` | | | Request header in `Key: Value` format (repeatable) |
| `--body` | | | Request body for the probe request |
| `--host` | | route first host | Host header for the probe request |
| `--gateway-url` | | derived from Admin URL (`:9080`) | APISIX data plane gateway URL |
| `--control-url` | | derived from Admin URL (`:9090`) | APISIX Control API URL |
| `--output` | `-o` | `table` for TTY, `json` for non-TTY | Output format (`table`, `json`, `yaml`) |

Gateway URL precedence:

1. `--gateway-url`
2. `APISIX_GATEWAY_URL`
3. Derived from current Admin API host using port `9080`

Control URL precedence:

1. `--control-url`
2. Derived from current Admin API host using port `9090`

### Examples

Basic trace:

```bash
a6 debug trace my-route
```

POST with custom path and headers:

```bash
a6 debug trace my-route \
  --method POST \
  --path /orders \
  --header "Content-Type: application/json" \
  --header "X-Debug: 1" \
  --body '{"order_id":"123"}'
```

JSON output for scripts:

```bash
a6 debug trace my-route -o json
```

Custom gateway and control URLs:

```bash
a6 debug trace my-route \
  --gateway-url http://127.0.0.1:9080 \
  --control-url http://127.0.0.1:9090
```

### Table Output

Typical table output includes:

- Route summary (ID, URI, methods, hosts, upstream)
- Probe request summary (method and URL)
- Gateway response status and latency
- Configured plugins ordered by execution priority
- Upstream status and executed plugins (if APISIX debug headers are enabled)

If APISIX does not return `Apisix-Plugins`, the output explains that APISIX debug mode should be enabled to expose executed plugins.

## `a6 debug logs`

Stream APISIX logs in real time. APISIX does not provide a native log-streaming API, so `a6` supports two modes:

- **Docker mode (default):** runs `docker logs` against the APISIX container
- **File mode:** tails a local log file when `--file` is provided

```bash
a6 debug logs
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--follow` | `-f` | `false` | Stream logs continuously |
| `--tail` | `-n` | `100` | Number of recent lines to show |
| `--since` | | `""` | Show logs since duration (e.g., `5m`, `1h`, `24h`) |
| `--type` | `-t` | `all` | Log type: `error`, `access`, `all` |
| `--container` | `-c` | auto-detect | Docker container name |
| `--file` | | `""` | Path to log file (use file tailing instead of Docker) |
| `--output` | `-o` | `""` | Output format option (logs are passed through as raw lines) |

### Examples

Stream the last 100 lines from auto-detected APISIX container:

```bash
a6 debug logs
```

Follow logs from a specific container:

```bash
a6 debug logs --container apisix --follow
```

Read logs from the last hour:

```bash
a6 debug logs --container apisix --since 1h --tail 200
```

Tail a local file:

```bash
a6 debug logs --file /usr/local/apisix/logs/error.log --tail 200
```

Follow a local file continuously:

```bash
a6 debug logs --file /usr/local/apisix/logs/access.log --follow
```
