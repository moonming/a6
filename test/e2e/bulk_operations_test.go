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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createLabeledRouteViaCLI(t *testing.T, env []string, id, uri, labelKey, labelValue string) {
	t.Helper()
	routeJSON := fmt.Sprintf(`{"id":"%s","uri":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"%s":"%s"}}`,
		id, uri, labelKey, labelValue)
	tmpFile := filepath.Join(t.TempDir(), id+".json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(routeJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "route create %s failed: stdout=%s stderr=%s", id, stdout, stderr)
}

func routeExistsViaCLI(t *testing.T, env []string, id string) bool {
	t.Helper()
	_, _, err := runA6WithEnv(env, "route", "get", id)
	return err == nil
}

func deleteRouteViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	runA6WithEnv(env, "route", "delete", id, "--force")
}

func routeExistsViaAdmin(t *testing.T, id string) bool {
	t.Helper()
	resp, err := adminAPI(http.MethodGet, "/apisix/admin/routes/"+id, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func deleteRouteViaAdmin(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI(http.MethodDelete, "/apisix/admin/routes/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupBulkEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestBulkDeleteByLabel(t *testing.T) {
	id1 := "bulk-del-label-test-1"
	id2 := "bulk-del-label-test-2"
	id3 := "bulk-del-label-prod"

	env := setupBulkEnv(t)

	deleteRouteViaCLI(t, env, id1)
	deleteRouteViaCLI(t, env, id2)
	deleteRouteViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, id1)
		deleteRouteViaAdmin(t, id2)
		deleteRouteViaAdmin(t, id3)
	})

	createLabeledRouteViaCLI(t, env, id1, "/bulk/label/1", "env", "test")
	createLabeledRouteViaCLI(t, env, id2, "/bulk/label/2", "env", "test")
	createLabeledRouteViaCLI(t, env, id3, "/bulk/label/3", "env", "prod")

	stdout, stderr, err := runA6WithEnv(env, "route", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	assert.False(t, routeExistsViaAdmin(t, id1))
	assert.False(t, routeExistsViaAdmin(t, id2))
	assert.True(t, routeExistsViaAdmin(t, id3))
}

func TestBulkDeleteAll(t *testing.T) {
	id1 := "bulk-del-all-1"
	id2 := "bulk-del-all-2"

	env := setupBulkEnv(t)

	deleteRouteViaCLI(t, env, id1)
	deleteRouteViaCLI(t, env, id2)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, id1)
		deleteRouteViaAdmin(t, id2)
	})

	createLabeledRouteViaCLI(t, env, id1, "/bulk/all/1", "suite", "bulk-all")
	createLabeledRouteViaCLI(t, env, id2, "/bulk/all/2", "suite", "bulk-all")

	stdout, stderr, err := runA6WithEnv(env, "route", "delete", "--all", "--force")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	assert.False(t, routeExistsViaAdmin(t, id1))
	assert.False(t, routeExistsViaAdmin(t, id2))
}

func TestBulkExportByLabel(t *testing.T) {
	id1 := "bulk-export-1"
	id2 := "bulk-export-2"
	id3 := "bulk-export-nolabel"

	env := setupBulkEnv(t)

	deleteRouteViaCLI(t, env, id1)
	deleteRouteViaCLI(t, env, id2)
	deleteRouteViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, id1)
		deleteRouteViaAdmin(t, id2)
		deleteRouteViaAdmin(t, id3)
	})

	createLabeledRouteViaCLI(t, env, id1, "/bulk/export/1", "env", "staging")
	createLabeledRouteViaCLI(t, env, id2, "/bulk/export/2", "env", "staging")
	createLabeledRouteViaCLI(t, env, id3, "/bulk/export/3", "env", "prod")

	stdout, stderr, err := runA6WithEnv(env, "route", "export", "--label", "env=staging", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	var routes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &routes), "stdout=%s", stdout)
	assert.Len(t, routes, 2)

	ids := make([]string, 0, 2)
	for _, r := range routes {
		if id, ok := r["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	joined := strings.Join(ids, ",")
	assert.Contains(t, joined, id1)
	assert.Contains(t, joined, id2)
	assert.NotContains(t, joined, id3, "route with env=prod should not be exported when filtering by env=staging")
}
