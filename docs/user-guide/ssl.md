# SSL Certificate Management

The `a6 ssl` command allows you to manage Apache APISIX SSL certificates. You can list, create, update, get, and delete SSL certificates using the CLI.

## Commands

### `a6 ssl list`

Lists all SSL certificates in the current context.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--page` | | `1` | Page number for pagination |
| `--page-size` | | `20` | Number of items per page |
| `--output` | `-o` | `table` | Output format (table, json, yaml) |

**Examples:**

List all SSL certificates:
```bash
a6 ssl list
```

Output in JSON format:
```bash
a6 ssl list -o json
```

Paginated output:
```bash
a6 ssl list --page 2 --page-size 5
```

**Table Columns:**

| Column | Description |
|--------|-------------|
| ID | Unique identifier for the SSL certificate |
| SNI | Server Name Indication (comma-separated if multiple) |
| STATUS | Certificate status: enabled (1) or disabled (0) |
| TYPE | Certificate type (default: server) |
| VALIDITY | Certificate expiration date (parsed from PEM) |
| CREATED | Creation timestamp |

### `a6 ssl get`

Gets detailed information about a specific SSL certificate by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `yaml` | Output format (json, yaml) |

**Examples:**

Get SSL certificate by ID:
```bash
a6 ssl get 1
```

Get SSL certificate in JSON format:
```bash
a6 ssl get 1 -o json
```

### `a6 ssl create`

Creates a new SSL certificate from a JSON or YAML file. If the file contains an `id` field, the certificate is created with that ID (PUT). Otherwise, a new ID is generated (POST).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the SSL configuration file (required) |

**Examples:**

Create an SSL certificate from a JSON file:
```bash
a6 ssl create -f ssl.json
```

**Sample `ssl.json`:**
```json
{
  "id": "1",
  "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "snis": ["example.com", "*.example.com"]
}
```

### `a6 ssl update`

Updates an existing SSL certificate using a configuration file.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | | Path to the SSL configuration file (required) |

**Examples:**

Update SSL certificate with ID `1`:
```bash
a6 ssl update 1 -f updated-ssl.json
```

### `a6 ssl delete`

Deletes an SSL certificate by its ID.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | | `false` | Skip confirmation prompt |

**Examples:**

Delete SSL certificate with confirmation:
```bash
a6 ssl delete 1
```

Delete SSL certificate without confirmation:
```bash
a6 ssl delete 1 --force
```

## SSL Configuration Reference

Key fields in the SSL configuration:

| Field | Description |
|-------|-------------|
| `id` | Unique identifier for the SSL certificate |
| `cert` | PEM-encoded server certificate |
| `key` | PEM-encoded private key |
| `certs` | Array of additional PEM-encoded certificates |
| `keys` | Array of additional PEM-encoded private keys |
| `sni` | Single Server Name Indication |
| `snis` | Array of Server Name Indications |
| `client` | mTLS client verification settings (`ca`, `depth`) |
| `type` | Certificate type: `server` (default) or `client` |
| `status` | Certificate status: `1` for enabled, `0` for disabled |
| `ssl_protocols` | Allowed SSL/TLS protocols |
| `labels` | Key-value labels for the certificate |

For the full schema and detailed field descriptions, refer to the [APISIX SSL Admin API documentation](https://apisix.apache.org/docs/apisix/admin-api/#ssl).

## Examples

### Basic SSL certificate

Create an SSL certificate for a single domain.

```json
{
  "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "snis": ["example.com"]
}
```

### Wildcard SSL certificate

Create an SSL certificate with wildcard SNI.

```json
{
  "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "snis": ["*.example.com", "example.com"]
}
```

### SSL with mTLS client verification

Configure mutual TLS with client certificate verification.

```json
{
  "cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
  "snis": ["secure.example.com"],
  "client": {
    "ca": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
    "depth": 2
  }
}
```
