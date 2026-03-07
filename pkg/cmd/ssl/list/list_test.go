package list

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/httpmock"
	"github.com/api7/a6/pkg/iostreams"
)

type mockConfig struct {
	baseURL string
}

func (m *mockConfig) BaseURL() string                                 { return m.baseURL }
func (m *mockConfig) APIKey() string                                  { return "" }
func (m *mockConfig) CurrentContext() string                          { return "test" }
func (m *mockConfig) Contexts() []config.Context                      { return nil }
func (m *mockConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (m *mockConfig) AddContext(ctx config.Context) error             { return nil }
func (m *mockConfig) RemoveContext(name string) error                 { return nil }
func (m *mockConfig) SetCurrentContext(name string) error             { return nil }
func (m *mockConfig) Save() error                                     { return nil }

var listBody = `{
	"total": 2,
	"list": [
		{
			"key": "/apisix/ssls/1",
			"value": {
				"id": "1",
				"snis": ["test.example.com"],
				"status": 1,
				"type": "server",
				"cert": "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJALHMqLmvsSKPMA0GCSqGSIb3DQEBCwUAMBMxETAPBgNVBAMMCHRl\nc3QuY29tMCAXDTI1MDEwMTAwMDAwMFoYDzIxMjUwMTAxMDAwMDAwWjATMREwDwYD\nVQQDDAh0ZXN0LmNvbTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96po4MFKF6H\nRNZmRqKCDJMRGMapMNB2PNRqJNmz5vENgfqflNjm2Gk7JDDCQm7oqGEk/EXAMPLE\nAgMBAAEwDQYJKoZIhvcNAQELBQADQQA6AAFMAAAAAAAAAAAAAAAAAAAAAAAAAAAA\nAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\n-----END CERTIFICATE-----",
				"key": "-----BEGIN RSA PRIVATE KEY-----\nfake-key-data\n-----END RSA PRIVATE KEY-----",
				"create_time": 1704067200
			}
		},
		{
			"key": "/apisix/ssls/2",
			"value": {
				"id": "2",
				"sni": "api.example.com",
				"status": 0,
				"type": "client",
				"cert": "fake-cert",
				"key": "fake-key",
				"create_time": 1704153600
			}
		}
	]
}`

func TestSSLList_Table(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "SNI")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "TYPE")
	assert.Contains(t, output, "VALIDITY")
	assert.Contains(t, output, "CREATED")
	assert.Contains(t, output, "test.example.com")
	assert.Contains(t, output, "api.example.com")
	assert.Contains(t, output, "enabled")
	assert.Contains(t, output, "disabled")
	assert.Contains(t, output, "server")
	assert.Contains(t, output, "client")
	reg.Verify(t)
}

func TestSSLList_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	var result []interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	reg.Verify(t)
}

func TestSSLList_YAML(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls", httpmock.JSONResponse(listBody))

	ios, _, stdout, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "test.example.com")
	assert.Contains(t, output, "api.example.com")
	reg.Verify(t)
}

func TestSSLList_Empty(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls", httpmock.JSONResponse(`{"total":0,"list":[]}`))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No SSL certificates found.")
	reg.Verify(t)
}

func TestSSLList_APIError(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/ssls", httpmock.StringResponse(403, `{"error_msg":"forbidden"}`))

	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	reg.Verify(t)
}
