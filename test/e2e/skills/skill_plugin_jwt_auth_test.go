//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginJWTAuth(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-jwt-route"
	const username = "skill-jwt-user"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", username, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })
	t.Cleanup(func() { cleanupConsumer(t, username) })

	consumerJSON := `{"username":"skill-jwt-user","plugins":{"jwt-auth":{"key":"skill-jwt-key","secret":"my-secret-for-jwt-test-1234567890"}}}`
	consumerFile := writeJSON(t, "consumer", consumerJSON)
	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
	require.NoError(t, err, "consumer create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "consumer", "get", username, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"jwt-auth"`)

	routeJSON := `{"id":"skill-jwt-route","uri":"/skill-jwt-auth","plugins":{"jwt-auth":{}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err = runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"jwt-auth"`)
}
