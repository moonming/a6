//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteSecret(t *testing.T, managerID string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/secrets/"+managerID, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupSecretEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestSecret_CRUD(t *testing.T) {
	const secretID = "vault/test-vault-secret-crud-1"

	deleteSecret(t, secretID)
	t.Cleanup(func() { deleteSecret(t, secretID) })

	env := setupSecretEnv(t)

	createJSON := `{
		"uri": "http://127.0.0.1:8200",
		"prefix": "/apisix/kv",
		"token": "test-token-12345"
	}`
	createFile := filepath.Join(t.TempDir(), "secret.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "secret", "create", secretID, "-f", createFile)
	require.NoError(t, err, "secret create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, "test-vault-secret-crud-1")

	stdout, stderr, err = runA6WithEnv(env, "secret", "get", secretID)
	require.NoError(t, err, "secret get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "http://127.0.0.1:8200")

	stdout, stderr, err = runA6WithEnv(env, "secret", "list")
	require.NoError(t, err, "secret list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "vault/test-vault-secret-crud-1")

	updateJSON := `{
		"prefix": "/apisix/kv/updated",
		"token": "test-token-updated"
	}`
	updateFile := filepath.Join(t.TempDir(), "secret-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "secret", "update", secretID, "-f", updateFile)
	require.NoError(t, err, "secret update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "secret", "get", secretID)
	require.NoError(t, err, "secret get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "/apisix/kv/updated")

	stdout, stderr, err = runA6WithEnv(env, "secret", "delete", secretID, "--force")
	require.NoError(t, err, "secret delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "secret", "get", secretID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestSecret_GetNonExistent(t *testing.T) {
	env := setupSecretEnv(t)

	_, stderr, err := runA6WithEnv(env, "secret", "get", "vault/nonexistent-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestSecret_JSONOutput(t *testing.T) {
	const secretID = "vault/test-vault-secret-json-1"

	deleteSecret(t, secretID)
	t.Cleanup(func() { deleteSecret(t, secretID) })

	env := setupSecretEnv(t)

	createJSON := `{
		"uri": "http://127.0.0.1:8200",
		"prefix": "/apisix/kv",
		"token": "test-token-12345"
	}`
	createFile := filepath.Join(t.TempDir(), "secret-json.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "secret", "create", secretID, "-f", createFile)
	require.NoError(t, err, "secret create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "secret", "list", "--output", "json")
	require.NoError(t, err, "secret list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "test-vault-secret-json-1")

	stdout, stderr, err = runA6WithEnv(env, "secret", "get", secretID, "--output", "json")
	require.NoError(t, err, "secret get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "test-vault-secret-json-1")
}
