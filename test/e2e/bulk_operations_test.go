//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createLabeledRoute(t *testing.T, id, uri, labelKey, labelValue string) {
	t.Helper()
	body := fmt.Sprintf(`{"uri":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"%s":"%s"}}`, uri, labelKey, labelValue)
	resp, err := adminAPI(http.MethodPut, "/apisix/admin/routes/"+id, []byte(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)
}

func routeExists(t *testing.T, id string) bool {
	t.Helper()
	resp, err := adminAPI(http.MethodGet, "/apisix/admin/routes/"+id, nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func deleteRouteQuiet(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI(http.MethodDelete, "/apisix/admin/routes/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func TestBulkDeleteByLabel(t *testing.T) {
	id1 := "bulk-del-label-test-1"
	id2 := "bulk-del-label-test-2"
	id3 := "bulk-del-label-prod"

	deleteRouteQuiet(t, id1)
	deleteRouteQuiet(t, id2)
	deleteRouteQuiet(t, id3)
	t.Cleanup(func() {
		deleteRouteQuiet(t, id1)
		deleteRouteQuiet(t, id2)
		deleteRouteQuiet(t, id3)
	})

	createLabeledRoute(t, id1, "/bulk/label/1", "env", "test")
	createLabeledRoute(t, id2, "/bulk/label/2", "env", "test")
	createLabeledRoute(t, id3, "/bulk/label/3", "env", "prod")

	stdout, stderr, err := runA6("route", "delete", "--label", "env=test", "--force", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	assert.False(t, routeExists(t, id1))
	assert.False(t, routeExists(t, id2))
	assert.True(t, routeExists(t, id3))
}

func TestBulkDeleteAll(t *testing.T) {
	id1 := "bulk-del-all-1"
	id2 := "bulk-del-all-2"

	deleteRouteQuiet(t, id1)
	deleteRouteQuiet(t, id2)
	t.Cleanup(func() {
		deleteRouteQuiet(t, id1)
		deleteRouteQuiet(t, id2)
	})

	createLabeledRoute(t, id1, "/bulk/all/1", "suite", "bulk-all")
	createLabeledRoute(t, id2, "/bulk/all/2", "suite", "bulk-all")

	stdout, stderr, err := runA6("route", "delete", "--all", "--force", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	assert.False(t, routeExists(t, id1))
	assert.False(t, routeExists(t, id2))
}

func TestBulkExport(t *testing.T) {
	id1 := "bulk-export-1"
	id2 := "bulk-export-2"

	deleteRouteQuiet(t, id1)
	deleteRouteQuiet(t, id2)
	t.Cleanup(func() {
		deleteRouteQuiet(t, id1)
		deleteRouteQuiet(t, id2)
	})

	createLabeledRoute(t, id1, "/bulk/export/1", "env", "staging")
	createLabeledRoute(t, id2, "/bulk/export/2", "env", "staging")

	stdout, stderr, err := runA6("route", "export", "--label", "env=staging", "--output", "json", "--server", adminURL, "--api-key", adminKey)
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
}
