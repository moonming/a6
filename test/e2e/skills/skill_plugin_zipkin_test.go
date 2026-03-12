//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginZipkin verifies config-only setup of the zipkin plugin.
// Data-plane testing requires a Zipkin collector endpoint.
func TestSkillPluginZipkin(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-zipkin-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-zipkin-route",
		"uri": "/skill-zipkin",
		"plugins": {
			"zipkin": {
				"endpoint": "http://zipkin.example.com:9411/api/v2/spans",
				"sample_ratio": 1,
				"service_name": "skill-test"
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"zipkin"`)
	assert.Contains(t, stdout, `"skill-test"`)
}
