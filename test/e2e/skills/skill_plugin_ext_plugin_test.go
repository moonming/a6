//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginExtPlugin verifies config-only setup of the ext-plugin-pre-req plugin.
// Data-plane testing requires an external plugin runner process.
func TestSkillPluginExtPlugin(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ext-plugin-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-ext-plugin-route",
		"uri": "/skill-ext-plugin",
		"plugins": {
			"ext-plugin-pre-req": {
				"conf": [
					{"name": "skill-test-plugin", "value": "{\"enabled\":true}"}
				]
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"ext-plugin-pre-req"`)
	assert.Contains(t, stdout, `"skill-test-plugin"`)
}
