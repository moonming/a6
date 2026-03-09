package diff

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
	"github.com/api7/a6/pkg/cmdutil"
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
		"/apisix/admin/ssls",
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

func TestConfigDiff_CreateUpdateDelete(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 2,
		"list": [
			{"key":"/apisix/routes/r1","value":{"id":"r1","uri":"/a","name":"old"}},
			{"key":"/apisix/routes/r3","value":{"id":"r3","uri":"/c","name":"gone"}}
		]
	}`))

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /a
    name: new
  - id: r2
    uri: /b
    name: created
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.Error(t, err)
	assert.True(t, cmdutil.IsSilent(err))
	out := stdout.String()
	assert.Contains(t, out, "Differences found")
	assert.Contains(t, out, "routes: create=1 update=1 delete=1")
	assert.Contains(t, out, "CREATE r2")
	assert.Contains(t, out, "UPDATE r1")
	assert.Contains(t, out, "DELETE r3")
	reg.Verify(t)
}

func TestConfigDiff_NoDiff(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/routes/r1","value":{"id":"r1","uri":"/same","name":"same"}}]
	}`))

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /same
    name: same
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No differences found.")
	reg.Verify(t)
}

func TestConfigDiff_EmptyLocal(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/routes", httpmock.JSONResponse(`{
		"total": 1,
		"list": [{"key":"/apisix/routes/r1","value":{"id":"r1","uri":"/same","name":"same"}}]
	}`))

	local := writeConfig(t, `
version: "1"
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.Error(t, err)
	assert.True(t, cmdutil.IsSilent(err))
	assert.Contains(t, stdout.String(), "DELETE r1")
	reg.Verify(t)
}

func TestConfigDiff_EmptyRemote(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /same
    name: same
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.Error(t, err)
	assert.True(t, cmdutil.IsSilent(err))
	assert.Contains(t, stdout.String(), "CREATE r1")
	reg.Verify(t)
}

func TestConfigDiff_JSONOutput(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, nil)

	local := writeConfig(t, `
version: "1"
routes:
  - id: r1
    uri: /json
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local, "--output", "json"})
	err := c.Execute()

	require.Error(t, err)
	assert.True(t, cmdutil.IsSilent(err))

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	routes := result["routes"].(map[string]interface{})
	create := routes["create"].([]interface{})
	assert.Len(t, create, 1)
	reg.Verify(t)
}

func TestConfigDiff_StreamRoutesDisabled(t *testing.T) {
	reg := &httpmock.Registry{}
	registerEmptyResources(reg, map[string]bool{"/apisix/admin/stream_routes": true})
	reg.Register(http.MethodGet, "/apisix/admin/stream_routes", httpmock.StringResponse(http.StatusBadRequest,
		`{"error_msg":"stream mode is disabled, can not add stream routes"}`))

	local := writeConfig(t, `
version: "1"
`)

	ios, _, stdout, _ := iostreams.Test()
	c := NewCmdDiff(newFactory(reg, ios))
	c.SetArgs([]string{"-f", local})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No differences found.")
	reg.Verify(t)
}
