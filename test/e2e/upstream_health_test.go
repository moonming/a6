//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamHealth_WithHealthCheck(t *testing.T) {
	const (
		upstreamID = "test-upstream-health-1"
		routeID    = "test-route-health-1"
	)

	env := setupUpstreamEnv(t)

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID)
		deleteUpstreamViaAdmin(t, upstreamID)
	})

	upstreamBody := fmt.Sprintf(`{
		"id": "%s",
		"name": "health-test-upstream",
		"type": "roundrobin",
		"nodes": {"127.0.0.1:8080": 1},
		"checks": {
			"active": {
				"type": "http",
				"http_path": "/get",
				"healthy": {
					"interval": 1,
					"successes": 1
				},
				"unhealthy": {
					"interval": 1,
					"http_failures": 3
				}
			}
		}
	}`, upstreamID)
	upstreamFile := filepath.Join(t.TempDir(), "upstream-health-check.json")
	require.NoError(t, os.WriteFile(upstreamFile, []byte(upstreamBody), 0o644))
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", upstreamFile)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

	routeBody := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/test-health/*",
		"name": "health-test-route",
		"upstream_id": "%s",
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/test-health/(.*)", "/$1"]
			}
		}
	}`, routeID, upstreamID)
	routeFile := filepath.Join(t.TempDir(), "route-health-check.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeBody), 0o644))
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)

	time.Sleep(3 * time.Second)
	for i := 0; i < 5; i++ {
		resp, err := http.Get(gatewayURL + "/test-health/get")
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(3 * time.Second)

	stdout, stderr, err = runA6WithEnv(env, "upstream", "health", upstreamID,
		"--control-url", controlURL, "--output", "json")
	require.NoError(t, err, "upstream health failed: stdout=%s stderr=%s", stdout, stderr)

	var healthResp map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &healthResp))

	nodes, ok := healthResp["nodes"].([]interface{})
	require.True(t, ok, "nodes should be an array")
	require.NotEmpty(t, nodes, "should have at least one node")

	node := nodes[0].(map[string]interface{})
	assert.Equal(t, "127.0.0.1", node["ip"])
	assert.Equal(t, "healthy", node["status"])
}

func TestUpstreamHealth_NoHealthCheck(t *testing.T) {
	const upstreamID = "test-upstream-no-health-1"

	env := setupUpstreamEnv(t)

	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	t.Cleanup(func() { deleteUpstreamViaAdmin(t, upstreamID) })

	body := fmt.Sprintf(`{"id":"%s","name":"no-health-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`, upstreamID)
	upstreamFile := filepath.Join(t.TempDir(), "upstream-no-health-check.json")
	require.NoError(t, os.WriteFile(upstreamFile, []byte(body), 0o644))
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", upstreamFile)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "upstream", "health", upstreamID,
		"--control-url", controlURL)
	assert.Error(t, err, "should fail when no health check configured")
	assert.Contains(t, stderr, "health check", "error should mention health check")
}
