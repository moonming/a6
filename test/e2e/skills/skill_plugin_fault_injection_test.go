//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginFaultInjection(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-fault-inject-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-fault-inject-route","uri":"/skill-fault-inject","plugins":{"fault-injection":{"abort":{"http_status":503,"body":"fault injected","percentage":100}}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"fault-injection"`)

	status, body := httpGetWithRetry(t, gatewayURL+"/skill-fault-inject", nil, 503, 5*time.Second)
	assert.Equal(t, 503, status)
	assert.Contains(t, body, "fault injected")
}
