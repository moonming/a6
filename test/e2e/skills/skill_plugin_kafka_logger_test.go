//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginKafkaLogger verifies config-only setup of the kafka-logger plugin.
// Data-plane testing requires a Kafka broker.
func TestSkillPluginKafkaLogger(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-kafka-logger-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-kafka-logger-route",
		"uri": "/skill-kafka-logger",
		"plugins": {
			"kafka-logger": {
				"broker_list": {"127.0.0.1": 9092},
				"kafka_topic": "skill-test-topic",
				"batch_max_size": 1
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"kafka-logger"`)
	assert.Contains(t, stdout, `"skill-test-topic"`)
}
