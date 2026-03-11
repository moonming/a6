//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteSSLViaAdmin(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/ssls/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupSSLEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func readTestCert(t *testing.T) (string, string) {
	t.Helper()
	modRoot, err := resolveModuleRoot()
	require.NoError(t, err)
	certBytes, err := os.ReadFile(filepath.Join(modRoot, "test/e2e/testdata/test.crt"))
	require.NoError(t, err)
	keyBytes, err := os.ReadFile(filepath.Join(modRoot, "test/e2e/testdata/test.key"))
	require.NoError(t, err)
	return string(certBytes), string(keyBytes)
}

func TestSSL_CRUD(t *testing.T) {
	const sslID = "test-ssl-crud-1"

	env := setupSSLEnv(t)
	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { deleteSSLViaAdmin(t, sslID) })

	cert, key := readTestCert(t)

	// Escape newlines for JSON embedding
	certJSON := strings.ReplaceAll(cert, "\n", "\\n")
	keyJSON := strings.ReplaceAll(key, "\n", "\\n")

	// 1. Create
	sslJSON := `{
	"id": "` + sslID + `",
	"cert": "` + certJSON + `",
	"key": "` + keyJSON + `",
	"snis": ["test.example.com"]
}`
	tmpFile := filepath.Join(t.TempDir(), "ssl.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(sslJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", tmpFile)
	require.NoError(t, err, "ssl create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, sslID, "create output should mention SSL ID")

	// 2. Get
	stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID)
	require.NoError(t, err, "ssl get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test.example.com", "get output should contain SNI")

	// 3. List
	stdout, stderr, err = runA6WithEnv(env, "ssl", "list")
	require.NoError(t, err, "ssl list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test.example.com", "list output should contain SNI")

	// 4. Update
	updatedJSON := `{
	"cert": "` + certJSON + `",
	"key": "` + keyJSON + `",
	"snis": ["test.example.com", "test2.example.com"]
}`
	updatedFile := filepath.Join(t.TempDir(), "ssl-updated.json")
	require.NoError(t, os.WriteFile(updatedFile, []byte(updatedJSON), 0644))

	stdout, stderr, err = runA6WithEnv(env, "ssl", "update", sslID, "-f", updatedFile)
	require.NoError(t, err, "ssl update failed: stdout=%s stderr=%s", stdout, stderr)

	// 5. Get again: verify updated data
	stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID)
	require.NoError(t, err, "ssl get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "test2.example.com", "get output should contain updated SNI")

	// 6. Delete
	stdout, stderr, err = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	require.NoError(t, err, "ssl delete failed: stdout=%s stderr=%s", stdout, stderr)

	// 7. Get again: verify not found
	_, stderr, err = runA6WithEnv(env, "ssl", "get", sslID)
	assert.Error(t, err, "get after delete should fail")
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestSSL_CreateFromFile(t *testing.T) {
	const sslID = "test-ssl-fromfile-1"

	env := setupSSLEnv(t)
	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { deleteSSLViaAdmin(t, sslID) })

	cert, key := readTestCert(t)

	certJSON := strings.ReplaceAll(cert, "\n", "\\n")
	keyJSON := strings.ReplaceAll(key, "\n", "\\n")

	sslJSON := `{
	"id": "` + sslID + `",
	"cert": "` + certJSON + `",
	"key": "` + keyJSON + `",
	"snis": ["fromfile.example.com"],
	"status": 1
}`
	tmpFile := filepath.Join(t.TempDir(), "ssl.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(sslJSON), 0644))

	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", tmpFile)
	require.NoError(t, err, "ssl create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, sslID)

	// Verify via get in JSON format
	stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID, "--output", "json")
	require.NoError(t, err, "ssl get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "fromfile.example.com")
}

func TestSSL_ListWithStatus(t *testing.T) {
	const sslID = "test-ssl-liststatus-1"

	env := setupSSLEnv(t)
	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { deleteSSLViaAdmin(t, sslID) })

	cert, key := readTestCert(t)

	certJSON := strings.ReplaceAll(cert, "\n", "\\n")
	keyJSON := strings.ReplaceAll(key, "\n", "\\n")

	sslJSON := `{
	"id": "` + sslID + `",
	"cert": "` + certJSON + `",
	"key": "` + keyJSON + `",
	"snis": ["status-test.example.com"],
	"status": 1
}`
	tmpFile := filepath.Join(t.TempDir(), "ssl-status.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(sslJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", tmpFile)
	require.NoError(t, err, "ssl create failed: stdout=%s stderr=%s", stdout, stderr)

	// List with JSON output and verify status field
	stdout, stderr, err = runA6WithEnv(env, "ssl", "list", "--output", "json")
	require.NoError(t, err, "ssl list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, `"status": 1`, "list should show status value")
	assert.Contains(t, stdout, "status-test.example.com", "list should contain the SNI")
}

func TestSSL_GetShowsSNI(t *testing.T) {
	const sslID = "test-ssl-getsni-1"

	env := setupSSLEnv(t)
	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { deleteSSLViaAdmin(t, sslID) })

	cert, key := readTestCert(t)

	certJSON := strings.ReplaceAll(cert, "\n", "\\n")
	keyJSON := strings.ReplaceAll(key, "\n", "\\n")

	sslJSON := `{
	"id": "` + sslID + `",
	"cert": "` + certJSON + `",
	"key": "` + keyJSON + `",
	"snis": ["sni1.example.com", "sni2.example.com"]
}`
	tmpFile := filepath.Join(t.TempDir(), "ssl-sni.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(sslJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", tmpFile)
	require.NoError(t, err, "ssl create failed: stdout=%s stderr=%s", stdout, stderr)

	// Get in YAML and verify SNIs present
	stdout, stderr, err = runA6WithEnv(env, "ssl", "get", sslID)
	require.NoError(t, err, "ssl get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "sni1.example.com", "get should contain first SNI")
	assert.Contains(t, stdout, "sni2.example.com", "get should contain second SNI")
}
