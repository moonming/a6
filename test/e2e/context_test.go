//go:build e2e

package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_CreateAndUse(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	stdout, _, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180", "--api-key", "test123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "created")

	stdout, _, err = runA6WithEnv(env, "context", "current")
	require.NoError(t, err)
	assert.Contains(t, strings.TrimSpace(stdout), "local")

	_, _, err = runA6WithEnv(env, "context", "create", "staging",
		"--server", "http://staging:9180")
	require.NoError(t, err)

	_, _, err = runA6WithEnv(env, "context", "use", "staging")
	require.NoError(t, err)

	stdout, _, err = runA6WithEnv(env, "context", "current")
	require.NoError(t, err)
	assert.Equal(t, "staging", strings.TrimSpace(stdout))
}

func TestContext_List(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, _, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180")
	require.NoError(t, err)

	_, _, err = runA6WithEnv(env, "context", "create", "staging",
		"--server", "http://staging:9180")
	require.NoError(t, err)

	stdout, _, err := runA6WithEnv(env, "context", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "local")
	assert.Contains(t, stdout, "staging")
	assert.Contains(t, stdout, "http://localhost:9180")
	assert.Contains(t, stdout, "http://staging:9180")

	stdout, _, err = runA6WithEnv(env, "context", "list", "--output", "json")
	require.NoError(t, err)
	assert.True(t, json.Valid([]byte(stdout)), "output should be valid JSON")
	assert.Contains(t, stdout, "local")
	assert.Contains(t, stdout, "staging")
}

func TestContext_Delete(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, _, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180")
	require.NoError(t, err)

	_, _, err = runA6WithEnv(env, "context", "create", "staging",
		"--server", "http://staging:9180")
	require.NoError(t, err)

	_, _, err = runA6WithEnv(env, "context", "delete", "staging", "--force")
	require.NoError(t, err)

	stdout, _, err := runA6WithEnv(env, "context", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "local")
	assert.NotContains(t, stdout, "staging")
}

func TestContext_CreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, _, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180")
	require.NoError(t, err)

	_, stderr, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180")
	assert.Error(t, err)
	assert.Contains(t, stderr, "already exists")
}

func TestContext_UseNonExistent(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, stderr, err := runA6WithEnv(env, "context", "use", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, stderr, "not found")
}

func TestContext_DeleteActive(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, _, err := runA6WithEnv(env, "context", "create", "local",
		"--server", "http://localhost:9180")
	require.NoError(t, err)

	_, _, err = runA6WithEnv(env, "context", "delete", "local", "--force")
	require.NoError(t, err)

	stdout, _, err := runA6WithEnv(env, "context", "list")
	require.NoError(t, err)
	trimmed := strings.TrimSpace(stdout)
	assert.True(t, trimmed == "" || strings.Contains(trimmed, "No contexts"),
		"list should show empty or no-contexts message, got: %s", trimmed)
}
