//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginResponseRewrite(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-resp-rewrite-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-resp-rewrite-route","uri":"/skill-resp-rewrite","plugins":{"response-rewrite":{"status_code":201,"body":"skill-response-rewrite-body"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"response-rewrite"`)

	status, body := httpGetWithRetry(t, gatewayURL+"/skill-resp-rewrite", nil, 201, 5*time.Second)
	assert.Equal(t, 201, status)
	assert.Contains(t, body, "skill-response-rewrite-body")
}
