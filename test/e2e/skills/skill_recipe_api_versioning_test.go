//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeAPIVersioning(t *testing.T) {
	env := setupEnv(t)
	const v1RouteID = "skill-api-v1-route"
	const v2RouteID = "skill-api-v2-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", v1RouteID, "--force")
	_, _, _ = runA6WithEnv(env, "route", "delete", v2RouteID, "--force")
	t.Cleanup(func() { cleanupRoute(t, v1RouteID) })
	t.Cleanup(func() { cleanupRoute(t, v2RouteID) })

	v1RouteJSON := `{
		"id": "skill-api-v1-route",
		"uri": "/v1/skill-version/*",
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/v1/skill-version/(.*)", "/get"]
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "v1-route", v1RouteJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "v1 route create: stdout=%s stderr=%s", stdout, stderr)

	v2RouteJSON := `{
		"id": "skill-api-v2-route",
		"uri": "/v2/skill-version/*",
		"plugins": {
			"proxy-rewrite": {
				"regex_uri": ["^/v2/skill-version/(.*)", "/get"]
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f = writeJSON(t, "v2-route", v2RouteJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "v2 route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", v1RouteID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"proxy-rewrite"`)

	stdout, _, err = runA6WithEnv(env, "route", "get", v2RouteID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"proxy-rewrite"`)

	status, _ := httpGetWithRetry(t, gatewayURL+"/v1/skill-version/test", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)

	status, _ = httpGetWithRetry(t, gatewayURL+"/v2/skill-version/test", nil, 200, 5*time.Second)
	assert.Equal(t, 200, status)
}
