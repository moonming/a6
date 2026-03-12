//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginDatadog verifies config-only setup of the datadog plugin.
// Data-plane testing requires a Datadog agent.
func TestSkillPluginDatadog(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-datadog-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-datadog-route",
		"uri": "/skill-datadog",
		"plugins": {
			"datadog": {
				"prefer_name": true
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"datadog"`)
}
