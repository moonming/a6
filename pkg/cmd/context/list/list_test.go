package list

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type mockConfig struct {
	contexts       []config.Context
	currentContext string
}

func (m *mockConfig) BaseURL() string        { return "" }
func (m *mockConfig) APIKey() string         { return "" }
func (m *mockConfig) CurrentContext() string { return m.currentContext }
func (m *mockConfig) Contexts() []config.Context {
	out := make([]config.Context, len(m.contexts))
	copy(out, m.contexts)
	return out
}
func (m *mockConfig) GetContext(name string) (*config.Context, error) {
	for _, c := range m.contexts {
		if c.Name == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("context %q not found", name)
}
func (m *mockConfig) AddContext(ctx config.Context) error { return nil }
func (m *mockConfig) RemoveContext(name string) error     { return nil }
func (m *mockConfig) SetCurrentContext(name string) error { return nil }
func (m *mockConfig) Save() error                         { return nil }

func TestList_Empty(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "No contexts configured")
}

func TestList_SingleContext(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "local", Server: "http://localhost:9180"}},
		currentContext: "local",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "SERVER")
	assert.Contains(t, output, "CURRENT")
	assert.Contains(t, output, "local")
	assert.Contains(t, output, "http://localhost:9180")
	assert.Contains(t, output, "*")
}

func TestList_MultipleContexts(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts: []config.Context{
			{Name: "dev", Server: "http://dev:9180"},
			{Name: "prod", Server: "http://prod:9180"},
		},
		currentContext: "prod",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdList(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "dev")
	assert.Contains(t, output, "prod")
	assert.Contains(t, output, "http://dev:9180")
	assert.Contains(t, output, "http://prod:9180")
}

func TestList_JSONOutput(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts: []config.Context{
			{Name: "dev", Server: "http://dev:9180"},
		},
		currentContext: "dev",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	// Need to create the command with the root flag available
	c := NewCmdList(f)
	// Simulate the --output flag inherited from root
	c.Flags().StringP("output", "o", "", "Output format")
	c.SetArgs([]string{"--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, `"name": "dev"`)
	assert.Contains(t, output, `"server": "http://dev:9180"`)
	assert.Contains(t, output, `"current": true`)
}

func TestList_YAMLOutput(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts: []config.Context{
			{Name: "prod", Server: "http://prod:9180"},
		},
		currentContext: "prod",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdList(f)
	c.Flags().StringP("output", "o", "", "Output format")
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "name: prod")
	assert.Contains(t, output, "server: http://prod:9180")
	assert.Contains(t, output, "current: true")
}
