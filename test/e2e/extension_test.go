//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionListEmpty(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	stdout, stderr, err := runA6WithEnv(env, "extension", "list")
	require.NoError(t, err, "extension list failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "No extensions installed.")
}

func TestExtensionInstallInvalidFormat(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, stderr, err := runA6WithEnv(env, "extension", "install", "bad-format")
	require.Error(t, err)
	assert.Contains(t, stderr, "owner/repo")
}

func TestExtensionRemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	env := []string{"A6_CONFIG_DIR=" + dir}

	_, stderr, err := runA6WithEnv(env, "extension", "remove", "nonexistent", "--force")
	require.Error(t, err)
	assert.Contains(t, stderr, "not found")
}

func TestExtensionHelp(t *testing.T) {
	stdout, stderr, err := runA6WithEnv([]string{"A6_CONFIG_DIR=" + t.TempDir()}, "extension", "--help")
	require.NoError(t, err, "extension --help failed: stderr=%s", stderr)
	assert.Contains(t, stdout, "install")
	assert.Contains(t, stdout, "list")
	assert.Contains(t, stdout, "upgrade")
	assert.Contains(t, stdout, "remove")
}

func TestExtensionAlias(t *testing.T) {
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	stdout1, stderr1, err1 := runA6WithEnv(env, "extension", "--help")
	require.NoError(t, err1, "extension --help failed: stderr=%s", stderr1)

	stdout2, stderr2, err2 := runA6WithEnv(env, "ext", "--help")
	require.NoError(t, err2, "ext --help failed: stderr=%s", stderr2)

	assert.Equal(t, stdout1, stdout2)
}
