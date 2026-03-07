package get

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

var singleCredentialBody = `{
	"key": "/apisix/consumers/jack/credentials/cred-1",
	"value": {
		"id": "cred-1",
		"plugins": {
			"key-auth": {
				"key": "test-key"
			}
		}
	}
}`

func TestCredentialGet_JSON(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers/jack/credentials/cred-1", httpmock.JSONResponse(singleCredentialBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"cred-1", "--consumer", "jack"})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "cred-1", result["id"])
	reg.Verify(t)
}

func TestCredentialGet_YAML(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers/jack/credentials/cred-1", httpmock.JSONResponse(singleCredentialBody))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"cred-1", "--consumer", "jack", "--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "id: cred-1")
	assert.Contains(t, stdout.String(), "key-auth")
	reg.Verify(t)
}

func TestCredentialGet_NotFound(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodGet, "/apisix/admin/consumers/jack/credentials/cred-999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"cred-999", "--consumer", "jack"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	reg.Verify(t)
}

func TestCredentialGet_MissingConsumer(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdGet(f)
	c.SetArgs([]string{"cred-1"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
