//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginWolfRBAC verifies config-only setup of the wolf-rbac plugin.
// Data-plane testing requires a Wolf RBAC server.
func TestSkillPluginWolfRBAC(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-wolf-rbac-route"
	const username = "skill-wolf-user"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", username, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupConsumer(t, username) })

	consumerJSON := `{
		"username": "skill-wolf-user",
		"plugins": {
			"wolf-rbac": {
				"server": "http://wolf.example.com",
				"appid": "test-app"
			}
		}
	}`
	f := writeJSON(t, "consumer", consumerJSON)
	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", f)
	require.NoError(t, err, "consumer create: stdout=%s stderr=%s", stdout, stderr)

	routeJSON := `{
		"id": "skill-wolf-rbac-route",
		"uri": "/skill-wolf-rbac",
		"plugins": {
			"wolf-rbac": {}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"wolf-rbac"`)
}
