//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteRouteViaAdmin and deleteRouteViaCLI live in bulk_operations_test.go.

func createTestRouteViaCLI(t *testing.T, env []string, id, name, uri string) {
	t.Helper()
	body := fmt.Sprintf(`{"id":"%s","uri":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`, id, uri, name)
	tmpFile := filepath.Join(t.TempDir(), id+".json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(body), 0644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "failed to create test route %s: stdout=%s stderr=%s", id, stdout, stderr)
}

// setupRouteEnv returns env vars and creates a context pointing at the real APISIX.
func setupRouteEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestRoute_CRUD(t *testing.T) {
	const routeID = "test-route-crud-1"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	// 1. Create: write a JSON file and run `a6 route create -f file.json`
	routeJSON := `{
	"id": "test-route-crud-1",
	"uri": "/test-crud",
	"name": "test-crud-route",
	"methods": ["GET"],
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "route.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(routeJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, routeID, "create output should mention route ID")

	// 2. Get: verify route data
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route", "get output should contain route name")
	assert.Contains(t, stdout, "/test-crud", "get output should contain route URI")

	// 3. List: verify the created route appears
	stdout, stderr, err = runA6WithEnv(env, "route", "list")
	require.NoError(t, err, "route list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route", "list output should contain route name")

	// 4. Update: write updated JSON and run `a6 route update {id} -f file.json`
	updatedJSON := `{
	"uri": "/test-crud-updated",
	"name": "test-crud-route-updated",
	"methods": ["GET", "POST"],
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	updatedFile := filepath.Join(t.TempDir(), "route-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "route", "update", routeID, "-f", updatedFile)
	require.NoError(t, err, "route update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-route-updated", "get output should contain updated name")
	assert.Contains(t, stdout, "/test-crud-updated", "get output should contain updated URI")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "route", "delete", routeID, "--force")
	require.NoError(t, err, "route delete failed: stdout=%s stderr=%s", stdout, stderr)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_ListEmpty(t *testing.T) {
	env := setupRouteEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "route", "list")
	require.NoError(t, err, "route list failed: stdout=%s stderr=%s", stdout, stderr)
	// When no routes exist, should show an empty result or a "no routes" message.
	combined := stdout + stderr
	noRoutes := strings.Contains(combined, "No routes") ||
		strings.Contains(combined, "no routes") ||
		strings.Contains(combined, "0 route") ||
		// An empty table (just headers) is also acceptable.
		!strings.Contains(combined, "/")
	assert.True(t, noRoutes || stdout == "" || strings.TrimSpace(stdout) == "",
		"list should indicate no routes found, got: %s", combined)
}

func TestRoute_ListWithFilters(t *testing.T) {
	const (
		routeID1 = "test-route-filter-1"
		routeID2 = "test-route-filter-2"
	)
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID1)
	deleteRouteViaCLI(t, env, routeID2)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID1)
		deleteRouteViaAdmin(t, routeID2)
	})

	createTestRouteViaCLI(t, env, routeID1, "route-one-alpha", "/filter-alpha")
	createTestRouteViaCLI(t, env, routeID2, "route-two-beta", "/filter-beta")

	// Filter by name — should show only the matching route.
	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--name", "route-one-alpha")
	require.NoError(t, err, "route list --name failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "route-one-alpha", "filtered list should contain matching route")
	assert.NotContains(t, stdout, "route-two-beta", "filtered list should not contain other route")
}

func TestRoute_GetNonExistent(t *testing.T) {
	env := setupRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "route", "get", "nonexistent-999")
	assert.Error(t, err, "get nonexistent route should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_DeleteNonExistent(t *testing.T) {
	env := setupRouteEnv(t)

	_, stderr, err := runA6WithEnv(env, "route", "delete", "nonexistent-999", "--force")
	assert.Error(t, err, "delete nonexistent route should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestRoute_DeleteWithForce(t *testing.T) {
	const routeID = "test-route-force-del"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "force-delete-route", "/force-delete")

	stdout, stderr, err := runA6WithEnv(env, "route", "delete", routeID, "--force")
	require.NoError(t, err, "route delete --force failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	assert.True(t, strings.Contains(combined, "deleted") || strings.Contains(combined, "Deleted") || strings.Contains(combined, routeID),
		"delete output should confirm deletion, got: %s", combined)

	// Verify route is gone via Admin API.
	resp, err := adminAPI("GET", "/apisix/admin/routes/"+routeID, nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode, "route should not exist after deletion")
}

func TestRoute_JSONOutput(t *testing.T) {
	const routeID = "test-route-json-out"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "json-output-route", "/json-output")

	// List with JSON output.
	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--output", "json")
	require.NoError(t, err, "route list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "json-output-route", "JSON list should contain route name")

	// Get with JSON output.
	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err, "route get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, routeID, "JSON get should contain route ID")
	assert.Contains(t, stdout, "json-output-route", "JSON get should contain route name")
}

func TestRoute_YAMLOutput(t *testing.T) {
	const routeID = "test-route-yaml-out"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	createTestRouteViaCLI(t, env, routeID, "yaml-output-route", "/yaml-output")

	stdout, stderr, err := runA6WithEnv(env, "route", "get", routeID, "--output", "yaml")
	require.NoError(t, err, "route get --output yaml failed: stdout=%s stderr=%s", stdout, stderr)
	// YAML output should contain typical YAML markers like key: value pairs.
	assert.Contains(t, stdout, "name:", "YAML output should contain 'name:' key")
	assert.Contains(t, stdout, "uri:", "YAML output should contain 'uri:' key")
	assert.Contains(t, stdout, "yaml-output-route", "YAML output should contain route name")
}

func TestRoute_TrafficForwarding(t *testing.T) {
	const routeID = "test-route-traffic"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	routeBody := fmt.Sprintf(`{
		"id": "%s",
		"uri": "/test-e2e-traffic/*",
		"name": "traffic-test-route",
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/test-e2e-traffic/(.*)", "/$1"]
			}
		},
		"upstream": {
			"type": "roundrobin",
			"nodes": {
				"127.0.0.1:8080": 1
			}
		}
	}`, routeID)
	tmpFile := filepath.Join(t.TempDir(), "route-traffic.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(routeBody), 0644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "failed to create traffic test route: stdout=%s stderr=%s", stdout, stderr)

	// Retry to allow APISIX to propagate the route config from etcd.
	var gwResp *http.Response
	for i := 0; i < 10; i++ {
		gwResp, err = http.Get(gatewayURL + "/test-e2e-traffic/get")
		require.NoError(t, err, "gateway request should succeed")
		if gwResp.StatusCode == http.StatusOK {
			break
		}
		gwResp.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}
	defer gwResp.Body.Close()
	assert.Equal(t, http.StatusOK, gwResp.StatusCode, "gateway should return 200 for proxied request")
}
