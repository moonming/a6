//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeGraphQLProxy(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-graphql-proxy-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-graphql-proxy-route",
		"uri": "/skill-graphql",
		"methods": ["POST"],
		"plugins": {
			"key-auth": {},
			"limit-count": {
				"count": 1000,
				"time_window": 60,
				"key_type": "var",
				"key": "consumer_name"
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"key-auth"`)
	assert.Contains(t, stdout, `"limit-count"`)
}
