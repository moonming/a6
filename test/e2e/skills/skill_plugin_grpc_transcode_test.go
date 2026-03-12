//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginGRPCTranscode verifies config-only setup of the grpc-transcode plugin.
// Data-plane testing requires a gRPC backend and a proto definition.
func TestSkillPluginGRPCTranscode(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-grpc-transcode-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-grpc-transcode-route",
		"uri": "/skill-grpc-transcode",
		"methods": ["GET", "POST"],
		"plugins": {
			"grpc-transcode": {
				"proto_id": "1",
				"service": "helloworld.Greeter",
				"method": "SayHello"
			}
		},
		"upstream": {"scheme": "grpc", "type": "roundrobin", "nodes": {"127.0.0.1:50051": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"grpc-transcode"`)
	assert.Contains(t, stdout, `"helloworld.Greeter"`)
}
