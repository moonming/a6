//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractive_RouteGetRequiresIDInNonTTY(t *testing.T) {
	_, stderr, err := runA6("route", "get", "--server", adminURL, "--api-key", adminKey)
	require.Error(t, err)
	assert.True(t, strings.Contains(stderr, "id argument is required") || strings.Contains(stderr, "no routes found"),
		"expected 'id argument is required' or 'no routes found', got: %s", stderr)
}

func TestInteractive_UpstreamHealthRequiresIDInNonTTY(t *testing.T) {
	_, stderr, err := runA6("upstream", "health", "--server", adminURL, "--api-key", adminKey)
	require.Error(t, err)
	assert.True(t, strings.Contains(stderr, "id argument is required") || strings.Contains(stderr, "no upstreams found"),
		"expected 'id argument is required' or 'no upstreams found', got: %s", stderr)
}

func TestInteractive_ExplicitIDStillWorks(t *testing.T) {
	const routeID = "test-interactive-explicit-id"

	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, routeID)
	t.Cleanup(func() { deleteRouteViaAdmin(t, routeID) })

	routeJSON := `{"id":"test-interactive-explicit-id","name":"interactive-explicit","uri":"/interactive-explicit","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	tmpFile := filepath.Join(t.TempDir(), "route.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(routeJSON), 0o644))
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", tmpFile)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get with explicit ID failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "interactive-explicit")
}
