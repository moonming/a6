package create

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

var createResp = `{
	"key": "/apisix/upstreams/test-ups-1",
	"value": {
		"id": "test-ups-1",
		"name": "my-upstream",
		"type": "roundrobin",
		"nodes": {"127.0.0.1:8080": 1}
	}
}`

func TestUpstreamCreate_WithID(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/upstreams/test-ups-1", httpmock.JSONResponse(createResp))

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "upstream.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"id":"test-ups-1","name":"my-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`), 0644))

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", tmpFile})
	err := c.Execute()

	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test-ups-1", result["id"])
	reg.Verify(t)
}

func TestUpstreamCreate_WithoutID(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPost, "/apisix/admin/upstreams", httpmock.JSONResponse(createResp))

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "upstream.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"name":"my-upstream","type":"roundrobin","nodes":{"127.0.0.1:8080":1}}`), 0644))

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", tmpFile})
	err := c.Execute()

	require.NoError(t, err)
	reg.Verify(t)
}

func TestUpstreamCreate_NoFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
