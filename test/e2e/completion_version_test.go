//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion_Bash(t *testing.T) {
	stdout, stderr, err := runA6("completion", "bash")
	require.NoError(t, err, "completion bash failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "bash completion", "should contain bash completion markers")
	assert.Contains(t, stdout, "__start_a6", "should contain the a6 completion function")
}

func TestCompletion_Zsh(t *testing.T) {
	stdout, stderr, err := runA6("completion", "zsh")
	require.NoError(t, err, "completion zsh failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "compdef", "should contain zsh compdef directive")
}

func TestVersion_Output(t *testing.T) {
	stdout, stderr, err := runA6("version")
	require.NoError(t, err, "version failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "a6 version", "should contain version header")
	assert.Contains(t, stdout, "commit:", "should contain commit info")
	assert.Contains(t, stdout, "go:", "should contain Go version")
	assert.Contains(t, stdout, "platform:", "should contain platform info")
}

func TestVersion_JSON(t *testing.T) {
	stdout, stderr, err := runA6("version", "--output", "json")
	require.NoError(t, err, "version --output json failed: stderr=%s", stderr)

	var info map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &info)
	require.NoError(t, err, "failed to parse version JSON: %s", stdout)
	assert.NotEmpty(t, info["version"])
	assert.NotEmpty(t, info["goVersion"])
	assert.NotEmpty(t, info["platform"])
}

func TestVersion_ContainsGoVersion(t *testing.T) {
	stdout, _, err := runA6("version")
	require.NoError(t, err)
	assert.True(t, strings.Contains(stdout, "go1."), "should contain Go version starting with go1.")
}
