package use

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
	saveErr        error
	saveCalled     bool
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
func (m *mockConfig) AddContext(ctx config.Context) error {
	m.contexts = append(m.contexts, ctx)
	return nil
}
func (m *mockConfig) RemoveContext(name string) error { return nil }
func (m *mockConfig) SetCurrentContext(name string) error {
	for _, c := range m.contexts {
		if c.Name == name {
			m.currentContext = name
			return nil
		}
	}
	return fmt.Errorf("context %q not found", name)
}
func (m *mockConfig) Save() error {
	m.saveCalled = true
	return m.saveErr
}

func TestUse_Success(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts: []config.Context{
			{Name: "dev", Server: "http://dev:9180"},
			{Name: "prod", Server: "http://prod:9180"},
		},
		currentContext: "dev",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdUse(f)
	c.SetArgs([]string{"prod"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Switched to context "prod"`)
	assert.Equal(t, "prod", cfg.currentContext)
	assert.True(t, cfg.saveCalled)
}

func TestUse_NonExistent(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "dev", Server: "http://dev:9180"}},
		currentContext: "dev",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdUse(f)
	c.SetArgs([]string{"nonexistent"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUse_SaveError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "dev", Server: "http://dev:9180"}},
		currentContext: "dev",
		saveErr:        fmt.Errorf("write failed"),
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdUse(f)
	c.SetArgs([]string{"dev"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}
