//go:build e2e

// Package skills provides per-skill e2e tests for a6 CLI.
package skills

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	binaryPath string
	adminURL   string
	gatewayURL string
	controlURL string
	httpbinURL string
	adminKey   = "edd1c9f034335f136f87ad84b625c8f1"
)

func TestMain(m *testing.M) {
	adminURL = envOrDefault("APISIX_ADMIN_URL", "http://127.0.0.1:9180")
	gatewayURL = envOrDefault("APISIX_GATEWAY_URL", "http://127.0.0.1:9080")
	controlURL = envOrDefault("APISIX_CONTROL_URL", "http://127.0.0.1:9090")
	httpbinURL = envOrDefault("HTTPBIN_URL", "http://127.0.0.1:8080")

	// Build the a6 binary into a temp directory.
	tmpDir, err := os.MkdirTemp("", "a6-e2e-skills-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	binaryPath = filepath.Join(tmpDir, "a6")

	modRoot, err := resolveModuleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve module root: %v\n", err)
		os.Exit(1)
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/a6")
	buildCmd.Dir = modRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build a6 binary: %v\n", err)
		os.Exit(1)
	}

	// Wait for APISIX Admin API to become healthy.
	if err := waitForHealthy(adminURL+"/apisix/admin/routes", 60*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "APISIX not ready: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// runA6WithEnv executes the a6 binary with custom environment variables.
func runA6WithEnv(env []string, args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), env...)
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// adminAPI sends an HTTP request to the APISIX Admin API.
// Used for test setup and cleanup only — not for testing the CLI itself.
func adminAPI(method, path string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, adminURL+path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, adminURL+path, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", adminKey)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

// setupEnv creates a fresh a6 context pointing at the real APISIX and returns
// the env slice to pass to runA6WithEnv.
func setupEnv(t *testing.T) []string {
	t.Helper()
	env := []string{
		"A6_CONFIG_DIR=" + t.TempDir(),
	}
	_, _, err := runA6WithEnv(env, "context", "create", "test",
		"--server", adminURL, "--api-key", adminKey)
	if err != nil {
		t.Fatalf("failed to create test context: %v", err)
	}
	return env
}

// writeJSON writes data to a temp JSON file and returns the path.
func writeJSON(t *testing.T, name, data string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name+".json")
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatalf("failed to write temp json: %v", err)
	}
	return p
}

// cleanupRoute deletes a route via Admin API (for t.Cleanup).
func cleanupRoute(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/routes/"+id, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

// cleanupUpstream deletes an upstream via Admin API.
func cleanupUpstream(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/upstreams/"+id, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

// cleanupConsumer deletes a consumer via Admin API.
func cleanupConsumer(t *testing.T, username string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/consumers/"+username, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

// cleanupService deletes a service via Admin API.
func cleanupService(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/services/"+id, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

// cleanupSSL deletes an SSL via Admin API.
func cleanupSSL(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/ssls/"+id, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

// cleanupGlobalRule deletes a global rule via Admin API.
func cleanupGlobalRule(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/global_rules/"+id, nil)
	if err == nil {
		_ = resp.Body.Close()
	}
}

func httpGet(t *testing.T, url string, headers map[string]string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http get failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

// httpGetWithRetry retries a GET request until expected status or timeout.
func httpGetWithRetry(t *testing.T, url string, headers map[string]string, expectedStatus int, timeout time.Duration) (int, string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastStatus int
	var lastBody string
	for time.Now().Before(deadline) {
		lastStatus, lastBody = httpGet(t, url, headers)
		if lastStatus == expectedStatus {
			return lastStatus, lastBody
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for status %d from %s, last status=%d body=%s",
		expectedStatus, url, lastStatus, lastBody)
	return lastStatus, lastBody
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func resolveModuleRoot() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env GOMOD: %w", err)
	}
	gomod := strings.TrimSpace(string(out))
	if gomod == "" || gomod == os.DevNull {
		return "", fmt.Errorf("not inside a Go module")
	}
	return filepath.Dir(gomod), nil
}

func waitForHealthy(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		req.Header.Set("X-API-KEY", adminKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 400 {
			return nil
		}
		lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for %s: %v", url, lastErr)
}
