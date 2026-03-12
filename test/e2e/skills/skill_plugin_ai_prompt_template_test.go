//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginAIPromptTemplate verifies config-only setup of the ai-prompt-template plugin.
// Data-plane testing requires access to an AI provider.
func TestSkillPluginAIPromptTemplate(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ai-prompt-template-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-ai-prompt-template-route",
		"uri": "/skill-ai-prompt-template",
		"methods": ["POST"],
		"plugins": {
			"ai-prompt-template": {
				"templates": [
					{
						"name": "skill-test-template",
						"template": {
							"model": "gpt-4",
							"messages": [
								{"role": "system", "content": "You are a helpful assistant."},
								{"role": "user", "content": "{{question}}"}
							]
						}
					}
				]
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"ai-prompt-template"`)
	assert.Contains(t, stdout, `"skill-test-template"`)
}
