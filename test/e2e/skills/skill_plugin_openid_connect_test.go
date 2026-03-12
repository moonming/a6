//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginOpenIDConnect verifies config-only setup of the openid-connect plugin.
// Data-plane testing requires an external OIDC provider (Keycloak, etc.).
func TestSkillPluginOpenIDConnect(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-oidc-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-oidc-route",
		"uri": "/skill-oidc",
		"plugins": {
			"openid-connect": {
				"client_id": "test-app",
				"client_secret": "test-secret",
				"discovery": "https://keycloak.example.com/realms/test/.well-known/openid-configuration",
				"bearer_only": true
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"openid-connect"`)
	assert.Contains(t, stdout, `"test-app"`)
}
