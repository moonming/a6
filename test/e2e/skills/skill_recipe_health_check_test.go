//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeHealthCheck(t *testing.T) {
	env := setupEnv(t)
	const upstreamID = "skill-health-check-upstream"
	const routeID = "skill-health-check-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "upstream", "delete", upstreamID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupUpstream(t, upstreamID) })

	upstreamJSON := `{
		"id": "skill-health-check-upstream",
		"type": "roundrobin",
		"nodes": {"127.0.0.1:8080": 1},
		"checks": {
			"active": {
				"type": "http",
				"http_path": "/get",
				"healthy": {
					"interval": 2,
					"successes": 2
				},
				"unhealthy": {
					"interval": 1,
					"http_failures": 3
				}
			}
		}
	}`
	f := writeJSON(t, "upstream", upstreamJSON)
	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", f)
	require.NoError(t, err, "upstream create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "upstream", "get", upstreamID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"checks"`)
	assert.Contains(t, stdout, `"http_path"`)

	routeJSON := `{"id":"skill-health-check-route","uri":"/skill-health-check","upstream_id":"skill-health-check-upstream"}`
	f = writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)
}
