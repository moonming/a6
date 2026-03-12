//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginAIProxy verifies config-only setup of the ai-proxy plugin.
// Data-plane testing requires access to an AI provider (OpenAI, etc.).
func TestSkillPluginAIProxy(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ai-proxy-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-ai-proxy-route",
		"uri": "/skill-ai-proxy",
		"methods": ["POST"],
		"plugins": {
			"ai-proxy": {
				"provider": "openai",
				"auth": {
					"header": {
						"Authorization": "Bearer sk-test-key-placeholder"
					}
				},
				"options": {
					"model": "gpt-4",
					"temperature": 0.7,
					"max_tokens": 1024
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
	assert.Contains(t, stdout, `"ai-proxy"`)
	assert.Contains(t, stdout, `"openai"`)
	assert.Contains(t, stdout, `"gpt-4"`)
}
