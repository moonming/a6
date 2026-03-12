//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeCircuitBreaker(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-circuit-breaker-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-circuit-breaker-route",
		"uri": "/skill-circuit-breaker",
		"plugins": {
			"api-breaker": {
				"break_response_code": 502,
				"unhealthy": {
					"http_statuses": [500, 502, 503],
					"failures": 3
				},
				"healthy": {
					"http_statuses": [200],
					"successes": 3
				}
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"api-breaker"`)
	assert.Contains(t, stdout, `"break_response_code"`)
}
