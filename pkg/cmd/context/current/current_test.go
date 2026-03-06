package current

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

func TestCurrent_HasContext(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "prod", Server: "http://prod:9180"}},
		currentContext: "prod",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCurrent(f)
	c.SetArgs([]string{})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, "prod\n", out.String())
}

func TestCurrent_NoContext(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCurrent(f)
	c.SetArgs([]string{})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current context configured")
}

func TestCurrent_ConfigError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return nil, fmt.Errorf("config corrupted")
		},
	}

	c := NewCmdCurrent(f)
	c.SetArgs([]string{})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config corrupted")
}
