//go:build e2e

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateHelp(t *testing.T) {
	stdout, stderr, err := runA6("update", "--help")
	require.NoError(t, err, "update --help failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "Update a6 to the latest release")
	assert.Contains(t, stdout, "--force")
}

func TestUpdateDevVersion(t *testing.T) {
	env := []string{
		"HTTPS_PROXY=http://127.0.0.1:1",
		"HTTP_PROXY=http://127.0.0.1:1",
	}
	_, stderr, err := runA6WithEnv(env, "update", "--force")
	require.Error(t, err)
	assert.Contains(t, stderr, "dev build")
}

func TestVersionForUpdate(t *testing.T) {
	stdout, stderr, err := runA6("version", "--output", "json")
	require.NoError(t, err, "version --output json failed: stderr=%s", stderr)

	var info map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &info))
	version, ok := info["version"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, version)
}
