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

func deleteServiceViaAdmin(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/services/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func deleteServiceViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	_, _, _ = runA6WithEnv(env, "service", "delete", id, "--force")
}

func createTestServiceViaCLI(t *testing.T, env []string, id, name string) {
	t.Helper()
	body := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`, id, name)
	tmpFile := filepath.Join(t.TempDir(), id+".json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(body), 0644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", tmpFile)
	require.NoError(t, err, "failed to create test service %s: stdout=%s stderr=%s", id, stdout, stderr)
}

func setupServiceEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestService_CRUD(t *testing.T) {
	const serviceID = "test-svc-crud-1"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, serviceID)
	t.Cleanup(func() { deleteServiceViaAdmin(t, serviceID) })

	// 1. Create
	serviceJSON := `{
	"id": "test-svc-crud-1",
	"name": "test-crud-service",
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "service.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(serviceJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", tmpFile)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, serviceID, "create output should mention service ID")

	// 2. Get
	stdout, stderr, err = runA6WithEnv(env, "service", "get", serviceID)
	require.NoError(t, err, "service get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-service", "get output should contain service name")

	// 3. List
	stdout, stderr, err = runA6WithEnv(env, "service", "list")
	require.NoError(t, err, "service list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-service", "list output should contain service name")

	// 4. Update
	updatedJSON := `{
	"name": "test-crud-service-updated",
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`
	updatedFile := filepath.Join(t.TempDir(), "service-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "service", "update", serviceID, "-f", updatedFile)
	require.NoError(t, err, "service update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "service", "get", serviceID)
	require.NoError(t, err, "service get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test-crud-service-updated", "get output should contain updated name")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "service", "delete", serviceID, "--force")
	require.NoError(t, err, "service delete failed: stdout=%s stderr=%s", stdout, stderr)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "service", "get", serviceID)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestService_WithUpstream(t *testing.T) {
	const serviceID = "test-svc-upstream-1"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, serviceID)
	t.Cleanup(func() { deleteServiceViaAdmin(t, serviceID) })

	serviceJSON := `{
	"id": "test-svc-upstream-1",
	"name": "upstream-service",
	"upstream": {
		"type": "roundrobin",
		"scheme": "http",
		"nodes": {
			"127.0.0.1:8080": 1
		},
		"timeout": {
			"connect": 6,
			"send": 6,
			"read": 6
		}
	}
}`
	tmpFile := filepath.Join(t.TempDir(), "service.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(serviceJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", tmpFile)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "service", "get", serviceID, "--output", "json")
	require.NoError(t, err, "service get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "upstream-service", "get output should contain service name")
	assert.Contains(t, stdout, "roundrobin", "get output should contain upstream type")
	assert.Contains(t, stdout, "127.0.0.1:8080", "get output should contain upstream node")
}

func TestService_RouteWithServiceID(t *testing.T) {
	const (
		serviceID = "test-svc-route-ref"
		routeID   = "test-route-svc-ref"
	)
	env := setupServiceEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	deleteServiceViaCLI(t, env, serviceID)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, routeID)
		deleteServiceViaAdmin(t, serviceID)
	})

	// Create service via a6
	serviceJSON := fmt.Sprintf(`{
	"id": "%s",
	"name": "ref-service",
	"upstream": {
		"type": "roundrobin",
		"nodes": {
			"127.0.0.1:8080": 1
		}
	}
}`, serviceID)
	tmpFile := filepath.Join(t.TempDir(), "service.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(serviceJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", tmpFile)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)

	routeBody := fmt.Sprintf(`{"id":"%s","uri":"/test-svc-route/*","name":"svc-ref-route","service_id":"%s","plugins":{"proxy-rewrite":{"regex_uri":["^/test-svc-route/(.*)","/$1"]}}}`, routeID, serviceID)
	routeFile := filepath.Join(t.TempDir(), "route-with-service-id.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeBody), 0644))

	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "failed to create route with service_id: stdout=%s stderr=%s", stdout, stderr)

	var gwResp *http.Response
	for i := 0; i < 10; i++ {
		gwResp, err = http.Get(gatewayURL + "/test-svc-route/get")
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

func TestService_DeleteNonExistent(t *testing.T) {
	env := setupServiceEnv(t)

	_, stderr, err := runA6WithEnv(env, "service", "delete", "nonexistent-svc-999", "--force")
	assert.Error(t, err, "delete nonexistent service should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestService_JSONOutput(t *testing.T) {
	const serviceID = "test-svc-json-out"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, serviceID)
	t.Cleanup(func() { deleteServiceViaAdmin(t, serviceID) })

	createTestServiceViaCLI(t, env, serviceID, "json-output-service")

	stdout, stderr, err := runA6WithEnv(env, "service", "list", "--output", "json")
	require.NoError(t, err, "service list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "json-output-service", "JSON list should contain service name")

	stdout, stderr, err = runA6WithEnv(env, "service", "get", serviceID, "--output", "json")
	require.NoError(t, err, "service get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, serviceID, "JSON get should contain service ID")
	assert.Contains(t, stdout, "json-output-service", "JSON get should contain service name")
}
