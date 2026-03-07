package create

import (
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

var createdCredentialBody = `{
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

func TestCredentialCreate_Success(t *testing.T) {
	reg := &httpmock.Registry{}
	reg.Register(http.MethodPut, "/apisix/admin/consumers/jack/credentials/cred-1", httpmock.JSONResponse(createdCredentialBody))

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credential.json")
	err := os.WriteFile(filePath, []byte(`{"id":"cred-1","plugins":{"key-auth":{"key":"test-key"}}}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"--consumer", "jack", "-f", filePath})
	err = c.Execute()

	require.NoError(t, err)
	reg.Verify(t)
}

func TestCredentialCreate_MissingConsumer(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credential.json")
	err := os.WriteFile(filePath, []byte(`{"id":"cred-1","plugins":{"key-auth":{"key":"test-key"}}}`), 0o644)
	require.NoError(t, err)

	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"-f", filePath})
	err = c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
