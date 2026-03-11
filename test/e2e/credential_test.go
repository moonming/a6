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

func setupCredentialEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func deleteCredentialViaAdmin(t *testing.T, consumer, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/consumers/"+consumer+"/credentials/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func deleteCredTestConsumerViaAdmin(t *testing.T, username string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/consumers/"+username, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func TestCredential_CRUD(t *testing.T) {
	const (
		consumer = "test-cred-consumer"
		credID   = "cred-1"
	)

	env := setupCredentialEnv(t)

	_, _, _ = runA6WithEnv(env, "credential", "delete", credID, "--consumer", consumer, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", consumer, "--force")
	t.Cleanup(func() {
		deleteCredentialViaAdmin(t, consumer, credID)
		deleteCredTestConsumerViaAdmin(t, consumer)
	})

	consumerJSON := `{
		"username": "test-cred-consumer",
		"desc": "test consumer"
	}`
	consumerFile := filepath.Join(t.TempDir(), "credential-consumer-create.json")
	require.NoError(t, os.WriteFile(consumerFile, []byte(consumerJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
	require.NoError(t, err, "consumer create failed: stdout=%s stderr=%s", stdout, stderr)

	createJSON := `{
		"id": "cred-1",
		"plugins": {
			"key-auth": {
				"key": "test-credential-key-12345"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "credential-create.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "credential", "create", "--consumer", consumer, "-f", createFile)
	require.NoError(t, err, "credential create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, credID)

	stdout, stderr, err = runA6WithEnv(env, "credential", "get", credID, "--consumer", consumer)
	require.NoError(t, err, "credential get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "key-auth")

	stdout, stderr, err = runA6WithEnv(env, "credential", "list", "--consumer", consumer)
	require.NoError(t, err, "credential list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, credID)

	updateJSON := `{
		"plugins": {
			"key-auth": {
				"key": "test-credential-key-updated"
			}
		}
	}`
	updateFile := filepath.Join(t.TempDir(), "credential-update.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "credential", "update", credID, "--consumer", consumer, "-f", updateFile)
	require.NoError(t, err, "credential update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, credID)

	stdout, stderr, err = runA6WithEnv(env, "credential", "delete", credID, "--consumer", consumer, "--force")
	require.NoError(t, err, "credential delete failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, "deleted")
}

func TestCredential_GetNonExistent(t *testing.T) {
	const consumer = "test-cred-nonexistent-consumer"

	env := setupCredentialEnv(t)

	_, _, _ = runA6WithEnv(env, "consumer", "delete", consumer, "--force")
	t.Cleanup(func() { deleteCredTestConsumerViaAdmin(t, consumer) })

	consumerJSON := `{
		"username": "test-cred-nonexistent-consumer",
		"desc": "test consumer"
	}`
	consumerFile := filepath.Join(t.TempDir(), "credential-nonexistent-consumer-create.json")
	require.NoError(t, os.WriteFile(consumerFile, []byte(consumerJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
	require.NoError(t, err, "consumer create failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "credential", "get", "nonexistent-cred-999", "--consumer", consumer)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestCredential_JSONOutput(t *testing.T) {
	const (
		consumer = "test-cred-json-consumer"
		credID   = "cred-1"
	)

	env := setupCredentialEnv(t)

	_, _, _ = runA6WithEnv(env, "credential", "delete", credID, "--consumer", consumer, "--force")
	_, _, _ = runA6WithEnv(env, "consumer", "delete", consumer, "--force")
	t.Cleanup(func() {
		deleteCredentialViaAdmin(t, consumer, credID)
		deleteCredTestConsumerViaAdmin(t, consumer)
	})

	consumerJSON := `{
		"username": "test-cred-json-consumer",
		"desc": "test consumer"
	}`
	consumerFile := filepath.Join(t.TempDir(), "credential-json-consumer-create.json")
	require.NoError(t, os.WriteFile(consumerFile, []byte(consumerJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", consumerFile)
	require.NoError(t, err, "consumer create failed: stdout=%s stderr=%s", stdout, stderr)

	createJSON := `{
		"id": "cred-1",
		"plugins": {
			"key-auth": {
				"key": "test-credential-key-json"
			}
		}
	}`
	createFile := filepath.Join(t.TempDir(), "credential-json-create.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "credential", "create", "--consumer", consumer, "-f", createFile)
	require.NoError(t, err, "credential create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "credential", "get", credID, "--consumer", consumer, "--output", "json")
	require.NoError(t, err, "credential get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, credID)

	stdout, stderr, err = runA6WithEnv(env, "credential", "list", "--consumer", consumer, "--output", "json")
	require.NoError(t, err, "credential list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should produce valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, credID)
}
