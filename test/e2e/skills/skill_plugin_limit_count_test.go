//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginLimitCount(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-limit-count-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-limit-count-route","uri":"/skill-limit-count","plugins":{"limit-count":{"count":3,"time_window":60,"rejected_code":429,"key_type":"var","key":"remote_addr"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"limit-count"`)

	for i := 0; i < 3; i++ {
		status, _ := httpGetWithRetry(t, gatewayURL+"/skill-limit-count", nil, 200, 5*time.Second)
		assert.Equal(t, 200, status)
	}
	status, _ := httpGetWithRetry(t, gatewayURL+"/skill-limit-count", nil, 429, 5*time.Second)
	assert.Equal(t, 429, status)
}
