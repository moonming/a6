package sync

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func registerEmptyResources(reg *httpmock.Registry, skip map[string]bool) {
	resources := []string{
		"/apisix/admin/routes",
		"/apisix/admin/services",
		"/apisix/admin/upstreams",
		"/apisix/admin/consumers",
		"/apisix/admin/ssl",
		"/apisix/admin/global_rules",
		"/apisix/admin/plugin_configs",
		"/apisix/admin/consumer_groups",
		"/apisix/admin/stream_routes",
		"/apisix/admin/protos",
		"/apisix/admin/secrets",
	}
	for _, path := range resources {
		if skip != nil && skip[path] {
			continue
		}
		reg.Register(http.MethodGet, path, httpmock.JSONResponse(`{"total":0,"list":[]}`))
	}
	if skip == nil || !skip["/apisix/admin/plugins/list"] {
		reg.Register(http.MethodGet, "/apisix/admin/plugins/list", httpmock.JSONResponse(`[]`))
	}
}

func newFactory(reg *httpmock.Registry, ios *iostreams.IOStreams) *cmd.Factory {
	return &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return reg.GetClient(), nil },
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o644))
	return file
}

func TestConfigSync_CreatesNewResources(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)
	reg.Register(http.MethodPut, "/apisix/admin/routes/r1", httpmock.JSONResponse(`{"key":"/apisix/routes/r1","value":{"id":"r1"}}`))

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /sync
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, reg.CallCount(http.MethodPut, "/apisix/admin/routes/r1"))
	assert.Contains(t, stdout.String(), "Sync completed")
	reg.Verify(t)
}

func TestConfigSync_UpdatesExistingResources(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total":1,
		"list":[{"key":"/apisix/routes/r1","value":{"id":"r1","uri":"/old","name":"old"}}]
	}`))
	reg.Register(http.MethodPut, "/apisix/admin/routes/r1", httpmock.JSONResponse(`{"key":"/apisix/routes/r1","value":{"id":"r1"}}`))

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /new
    name: new
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, reg.CallCount(http.MethodPut, "/apisix/admin/routes/r1"))
	assert.Contains(t, stdout.String(), "updated=1")
	reg.Verify(t)
}

func TestConfigSync_DeletesRemoteOnlyResources(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total":1,
		"list":[{"key":"/apisix/routes/r-del","value":{"id":"r-del","uri":"/gone"}}]
	}`))
	reg.Register(http.MethodDelete, "/apisix/admin/routes/r-del", httpmock.JSONResponse(`{"deleted":"true"}`))

	local := writeConfig(t, `
version: "1"
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 1, reg.CallCount(http.MethodDelete, "/apisix/admin/routes/r-del"))
	assert.Contains(t, stdout.String(), "deleted=1")
	reg.Verify(t)
}

func TestConfigSync_DryRunDoesNotApply(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /sync
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local, "--dry-run"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 0, reg.CallCount(http.MethodPut, "/apisix/admin/routes/r1"))
	assert.Contains(t, stdout.String(), "Differences found")
	assert.Contains(t, stdout.String(), "CREATE r1")
	reg.Verify(t)
}

func TestConfigSync_DeleteFalseSkipsDeletion(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total":1,
		"list":[{"key":"/apisix/routes/r-del","value":{"id":"r-del","uri":"/gone"}}]
	}`))

	local := writeConfig(t, `
version: "1"
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local, "--delete=false"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, 0, reg.CallCount(http.MethodDelete, "/apisix/admin/routes/r-del"))
	assert.Contains(t, stdout.String(), "deleted=0")
	reg.Verify(t)
}

func TestConfigSync_ValidationFailureStopsSync(t *testing.T) {
	reg := &httpmock.Registry{}
	ios, _, _, _ := iostreams.Test()

	local := writeConfig(t, `
version: "1"
routes:
  - id: bad-route
`)

	c := NewCmdSync(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "validation failed")
	assert.Contains(t, err.Error(), "either uri or uris is required")
}
