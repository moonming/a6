//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginAIPromptDecorator verifies config-only setup of the ai-prompt-decorator plugin.
// Data-plane testing requires access to an AI provider.
func TestSkillPluginAIPromptDecorator(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ai-prompt-decorator-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-ai-prompt-decorator-route",
		"uri": "/skill-ai-prompt-decorator",
		"methods": ["POST"],
		"plugins": {
			"ai-prompt-decorator": {
				"prepend": [
					{"role": "system", "content": "Always respond in JSON format."}
				],
				"append": [
					{"role": "system", "content": "Be concise."}
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
	assert.Contains(t, stdout, `"ai-prompt-decorator"`)
	assert.Contains(t, stdout, `"Always respond in JSON format."`)
}
