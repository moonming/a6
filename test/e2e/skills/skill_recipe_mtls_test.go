//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeMTLS(t *testing.T) {
	env := setupEnv(t)
	const sslID = "skill-mtls-ssl"

	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { cleanupSSL(t, sslID) })

	sslJSON := `{
		"id": "skill-mtls-ssl",
		"cert": "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJAJOg/2FsJGsTMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnNl\ncnZlcjAeFw0yNDA1MDExMjAwMDBaFw0yNTA1MDExMjAwMDBaMBExDzANBgNVBAMM\nBnNlcnZlcjBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96MFSzyzqFCPcbGaXjm\n0uy4E+M0Z3vsDjGPUWMlYtfBih9GbU4BsZaLG+GKTFsBp0LBpQRGn/brhQiHSMGn\nAgMBAAGjUDBOMB0GA1UdDgQWBBT+7LoXkPJgQf8Vi5P0BNxvJIFVBTAfBgNVHSME\nGDAWgBT+7LoXkPJgQf8Vi5P0BNxvJIFVBTAMBgNVHRMEBTADAQH/MA0GCSqGSIb3\nDQEBCwUAA0EAE/qv+OUl0LjJfDMSJAzVwvqIFDCtPeRR04bHijAqe1JhKUON+GTR\nS/QFAj9hBgR8MBsV3Kbb3jHGMZB7JMYOyA==\n-----END CERTIFICATE-----",
		"key": "-----BEGIN RSA PRIVATE KEY-----\nMIIBogIBAAJBALuj3owVLPLOoUI9xsZpeObS7LgT4zRne+wOMY9RYyVi18GKH0Zt\nTgGxlosb4YpMWwGnQsGlBEaf9uuFCIdIwacCAwEAAQJAIFbVb8HMw2rlrGR/w63+\nv3R3VJ9jfP8CDU2qf3LHvhBEWPu+YcWIhskFXdmPtDeJVbB8VdSfbSt3l/uHKfvf\nwQIhAOnp8u0RHDqPeyPS4fFOt+CPj9nGI6lTW6wkz8LkuIPvAiEA38oBdR0l9FOU\nsGaQkPOKkWv2gMZl7M/44O+PDCKtSLECIBoK2MVSmYB+X2Q2cjwJG5EpcC8IVQMD\nCPSajNUb3ZtHAiEAgHHCAoE7l/Z8GQKJ+EJlN1GnJEAzRl+GiUKb4B4kGECIGZt\nGiH4xRhOJ1VJBBh9gwXY0WkNFi8u1LFKz/P/X/Kx\n-----END RSA PRIVATE KEY-----",
		"snis": ["skill-mtls.example.com"]
	}`
	f := writeJSON(t, "ssl", sslJSON)
	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", f)
	require.NoError(t, err, "ssl create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "ssl", "get", sslID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"skill-mtls.example.com"`)
}
