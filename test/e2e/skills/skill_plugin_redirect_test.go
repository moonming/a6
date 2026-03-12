//go:build e2e

package skills

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginRedirect(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-redirect-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-redirect-route","uri":"/skill-redirect","plugins":{"redirect":{"uri":"/redirected","ret_code":301}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"redirect"`)

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequest(http.MethodGet, gatewayURL+"/skill-redirect", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 301, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "/redirected")
}
