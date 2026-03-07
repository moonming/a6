---
name: a6-plugin-prometheus
description: >-
  Skill for configuring the Apache APISIX prometheus plugin via the a6 CLI.
  Covers enabling Prometheus metrics export on routes and globally, exposed
  metrics (HTTP status, latency, bandwidth, upstream health, LLM tokens),
  custom labels, histogram buckets, and Grafana dashboard integration.
version: "1.0.0"
author: Apache APISIX Contributors
license: Apache-2.0
metadata:
  category: plugin
  apisix_version: ">=3.0.0"
  plugin_name: prometheus
  a6_commands:
    - a6 route create
    - a6 route update
    - a6 config sync
---

# a6-plugin-prometheus

## Overview

The `prometheus` plugin exposes APISIX metrics in Prometheus text format. It
tracks HTTP status codes, request latency, bandwidth, upstream health, etcd
status, and (since v3.15) LLM token usage. Prometheus scrapes the metrics
endpoint; Grafana visualizes them.

## When to Use

- Monitor request rates, error rates, and latency per route/service/consumer
- Track upstream health check status
- Observe LLM token consumption and time-to-first-token
- Build dashboards and alerts with Prometheus + Grafana

## Plugin Configuration Reference (Route/Service/Global Rule)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `prefer_name` | boolean | No | `false` | Use route/service name instead of ID in metric labels |

The plugin has minimal per-route config. Most configuration is global via
`plugin_attr` in APISIX `config.yaml`.

## Metrics Exported

### Core Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `apisix_http_status` | counter | HTTP status codes per route/service/consumer |
| `apisix_http_latency` | histogram | Request latency in ms (types: request, upstream, apisix) |
| `apisix_bandwidth` | counter | Bandwidth in bytes (types: ingress, egress) |
| `apisix_http_requests_total` | gauge | Total HTTP requests received |
| `apisix_nginx_http_current_connections` | gauge | Current connections by state |
| `apisix_upstream_status` | gauge | Upstream health (1=healthy, 0=unhealthy) |
| `apisix_etcd_reachable` | gauge | etcd reachability (1=reachable, 0=unreachable) |
| `apisix_etcd_modify_indexes` | gauge | etcd modification count |
| `apisix_node_info` | gauge | APISIX node hostname and version |
| `apisix_shared_dict_capacity_bytes` | gauge | Shared memory capacity |
| `apisix_shared_dict_free_space_bytes` | gauge | Shared memory free space |
| `apisix_stream_connection_total` | counter | TCP/UDP stream connections |

### LLM/AI Metrics (v3.15+)

| Metric | Type | Description |
|--------|------|-------------|
| `apisix_llm_latency` | histogram | LLM request latency |
| `apisix_llm_prompt_tokens` | counter | Prompt tokens consumed |
| `apisix_llm_completion_tokens` | counter | Completion tokens consumed |
| `apisix_llm_active_connections` | gauge | Active LLM connections |

### Latency Types

- **request**: Total time from first byte read to last byte sent
- **upstream**: Time waiting for upstream response
- **apisix**: `request - upstream` (APISIX processing overhead)

## Step-by-Step: Enable Prometheus Metrics

### 1. Enable on a route

```bash
a6 route create -f - <<'EOF'
{
  "id": "my-api",
  "uri": "/api/*",
  "plugins": {
    "prometheus": {
      "prefer_name": true
    }
  },
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "backend:8080": 1
    }
  }
}
EOF
```

### 2. Enable globally (all routes)

```bash
curl "$(a6 context current -o json | jq -r .server)/apisix/admin/global_rules" \
  -X PUT \
  -H "X-API-KEY: $(a6 context current -o json | jq -r .api_key)" \
  -d '{
    "id": "prometheus-global",
    "plugins": {
      "prometheus": {}
    }
  }'
```

### 3. Access metrics

Default endpoint: `http://127.0.0.1:9091/apisix/prometheus/metrics`

### 4. Configure Prometheus scrape

```yaml
# prometheus.yml
scrape_configs:
  - job_name: apisix
    scrape_interval: 15s
    static_configs:
      - targets: ['127.0.0.1:9091']
```

### 5. Import Grafana dashboard

Import dashboard ID **11719** from grafana.com for a pre-built APISIX
monitoring dashboard.

## Common Patterns

### Custom metric prefix and export port

Configure in APISIX `config.yaml` (not via Admin API):

```yaml
plugin_attr:
  prometheus:
    export_uri: /apisix/prometheus/metrics
    metric_prefix: apisix_
    enable_export_server: true
    export_addr:
      ip: 0.0.0.0
      port: 9091
```

### Extra labels on metrics

```yaml
plugin_attr:
  prometheus:
    metrics:
      http_status:
        extra_labels:
          - upstream_addr: $upstream_addr
      http_latency:
        extra_labels:
          - upstream_addr: $upstream_addr
      bandwidth:
        extra_labels:
          - upstream_addr: $upstream_addr
```

### Custom histogram buckets

```yaml
plugin_attr:
  prometheus:
    default_buckets:
      - 10
      - 50
      - 100
      - 200
      - 500
      - 1000
      - 5000
      - 30000
```

## Config Sync Example

```yaml
version: "1"
global_rules:
  - id: prometheus-global
    plugins:
      prometheus:
        prefer_name: true
routes:
  - id: my-api
    uri: /api/*
    upstream_id: my-upstream
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| No metrics at endpoint | Plugin not enabled | Add `prometheus: {}` to route or global_rules |
| Metrics port unreachable | `enable_export_server: false` | Set to `true` or use `public-api` plugin |
| Missing route labels | `prefer_name: false` and route has no name | Set `prefer_name: true` and name your routes |
| No LLM metrics | APISIX < 3.15 or ai-proxy not configured | Upgrade APISIX; ensure ai-proxy is on the route |
| High cardinality | Too many extra labels | Reduce `extra_labels`; avoid high-cardinality variables |
