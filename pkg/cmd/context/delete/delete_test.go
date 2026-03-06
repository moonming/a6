package delete

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
func (m *mockConfig) AddContext(ctx config.Context) error { return nil }
func (m *mockConfig) RemoveContext(name string) error {
	idx := -1
	for i, c := range m.contexts {
		if c.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("context %q not found", name)
	}
	m.contexts = append(m.contexts[:idx], m.contexts[idx+1:]...)
	if m.currentContext == name {
		m.currentContext = ""
		if len(m.contexts) > 0 {
			m.currentContext = m.contexts[0].Name
		}
	}
	return nil
}
func (m *mockConfig) SetCurrentContext(name string) error { return nil }
func (m *mockConfig) Save() error {
	m.saveCalled = true
	return m.saveErr
}

func TestDelete_Success(t *testing.T) {
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

	c := NewCmdDelete(f)
	c.SetArgs([]string{"prod"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Context "prod" deleted.`)
	assert.True(t, cfg.saveCalled)
	assert.Len(t, cfg.contexts, 1)
}

func TestDelete_NonExistent(t *testing.T) {
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

	c := NewCmdDelete(f)
	c.SetArgs([]string{"ghost"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDelete_ActiveContext_WithForce(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
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

	c := NewCmdDelete(f)
	c.SetArgs([]string{"dev", "--force"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Context "dev" deleted.`)
	assert.Empty(t, cfg.contexts)
}

func TestDelete_ActiveContext_NonTTY_NoPrompt(t *testing.T) {
	// Non-TTY stdin should proceed without prompt (no blocking)
	ios, _, out, _ := iostreams.Test()
	// Test() creates non-TTY IOStreams by default
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

	c := NewCmdDelete(f)
	c.SetArgs([]string{"dev"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Context "dev" deleted.`)
}

func TestDelete_SaveError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "dev", Server: "http://dev:9180"}},
		currentContext: "dev",
		saveErr:        fmt.Errorf("permission denied"),
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"dev"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
