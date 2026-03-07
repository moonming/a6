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

func deleteUpstream(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/upstreams/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func createTestUpstream(t *testing.T, id, name string) {
	t.Helper()
	body := fmt.Sprintf(`{"name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`, name)
	resp, err := adminAPI("PUT", "/apisix/admin/upstreams/"+id, []byte(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create test upstream")
}

func setupUpstreamEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestUpstream_CRUD(t *testing.T) {
	const upstreamID = "test-upstream-crud-1"

	deleteUpstream(t, upstreamID)
	t.Cleanup(func() { deleteUpstream(t, upstreamID) })

	env := setupUpstreamEnv(t)

	// 1. Create
	upstreamJSON := `{
	"id": "test-upstream-crud-1",
	"name": "test-crud-upstream",
	"type": "roundrobin",
	"nodes": {
		"127.0.0.1:8080": 1
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "upstream.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(upstreamJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", tmpFile)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, upstreamID, "create output should mention upstream ID")

	// 2. Get
	stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID)
	require.NoError(t, err, "upstream get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-upstream", "get output should contain upstream name")

	// 3. List
	stdout, stderr, err = runA6WithEnv(env, "upstream", "list")
	require.NoError(t, err, "upstream list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-upstream", "list output should contain upstream name")

	// 4. Update
	updatedJSON := `{
	"name": "test-crud-upstream-updated",
	"type": "roundrobin",
	"nodes": {
		"127.0.0.1:9090": 1
	}
}`
	updatedFile := filepath.Join(t.TempDir(), "upstream-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "upstream", "update", upstreamID, "-f", updatedFile)
	require.NoError(t, err, "upstream update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID)
	require.NoError(t, err, "upstream get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-upstream-updated", "get output should contain updated name")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	require.NoError(t, err, "upstream delete failed: stdout=%s stderr=%s", stdout, stderr)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "upstream", "get", upstreamID)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestUpstream_CreateWithNodes(t *testing.T) {
	const upstreamID = "test-upstream-nodes-1"

	deleteUpstream(t, upstreamID)
	t.Cleanup(func() { deleteUpstream(t, upstreamID) })

	env := setupUpstreamEnv(t)

	// Create upstream with multiple nodes via Admin API
	body := `{"name":"multi-node-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1,"127.0.0.1:8081":2}}`
	resp, err := adminAPI("PUT", "/apisix/admin/upstreams/"+upstreamID, []byte(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)

	// Verify via a6 upstream get
	stdout, stderr, err := runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
	require.NoError(t, err, "upstream get failed: stdout=%s stderr=%s", stdout, stderr)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &result))
	assert.Equal(t, "multi-node-upstream", result["name"])

	nodes, ok := result["nodes"].(map[string]interface{})
	require.True(t, ok, "nodes should be a map")
	assert.Len(t, nodes, 2)
}

func TestUpstream_ListWithFilters(t *testing.T) {
	const (
		upstreamID1 = "test-upstream-filter-1"
		upstreamID2 = "test-upstream-filter-2"
	)

	deleteUpstream(t, upstreamID1)
	deleteUpstream(t, upstreamID2)
	t.Cleanup(func() {
		deleteUpstream(t, upstreamID1)
		deleteUpstream(t, upstreamID2)
	})

	createTestUpstream(t, upstreamID1, "upstream-one-alpha")
	createTestUpstream(t, upstreamID2, "upstream-two-beta")

	env := setupUpstreamEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "upstream", "list", "--name", "upstream-one-alpha")
	require.NoError(t, err, "upstream list --name failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "upstream-one-alpha", "filtered list should contain matching upstream")
	assert.NotContains(t, stdout, "upstream-two-beta", "filtered list should not contain other upstream")
}

func TestUpstream_DeleteNonExistent(t *testing.T) {
	env := setupUpstreamEnv(t)

	_, stderr, err := runA6WithEnv(env, "upstream", "delete", "nonexistent-999", "--force")
	assert.Error(t, err, "delete nonexistent upstream should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestUpstream_RouteWithUpstreamID(t *testing.T) {
	const (
		upstreamID = "test-upstream-route-ref"
		routeID    = "test-route-ups-ref"
	)

	deleteRoute(t, routeID)
	deleteUpstream(t, upstreamID)
	t.Cleanup(func() {
		deleteRoute(t, routeID)
		deleteUpstream(t, upstreamID)
	})

	env := setupUpstreamEnv(t)

	// Create upstream via a6
	upstreamJSON := fmt.Sprintf(`{
	"id": "%s",
	"name": "ref-upstream",
	"type": "roundrobin",
	"nodes": {
		"127.0.0.1:8080": 1
	}
}`, upstreamID)
	tmpFile := filepath.Join(t.TempDir(), "upstream.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(upstreamJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", tmpFile)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

	// Create route via Admin API with upstream_id.
	// proxy-rewrite strips the prefix so httpbin receives /get instead of /test-ups-ref/get.
	routeBody := fmt.Sprintf(`{"uri":"/test-ups-ref/*","name":"route-with-upstream-id","upstream_id":"%s","plugins":{"proxy-rewrite":{"regex_uri":["^/test-ups-ref/(.*)","/$1"]}}}`, upstreamID)
	resp, err := adminAPI("PUT", "/apisix/admin/routes/"+routeID, []byte(routeBody))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400, "failed to create route with upstream_id")

	var gwResp *http.Response
	for i := 0; i < 10; i++ {
		gwResp, err = http.Get(gatewayURL + "/test-ups-ref/get")
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
