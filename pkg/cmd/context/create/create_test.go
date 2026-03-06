package create

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

func (m *mockConfig) BaseURL() string {
	for _, c := range m.contexts {
		if c.Name == m.currentContext {
			return c.Server
		}
	}
	return ""
}

func (m *mockConfig) APIKey() string {
	for _, c := range m.contexts {
		if c.Name == m.currentContext {
			return c.APIKey
		}
	}
	return ""
}

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
	for _, c := range m.contexts {
		if c.Name == ctx.Name {
			return fmt.Errorf("context %q already exists", ctx.Name)
		}
	}
	m.contexts = append(m.contexts, ctx)
	if m.currentContext == "" {
		m.currentContext = ctx.Name
	}
	return nil
}

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

func TestCreate_Success(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"local", "--server", "http://localhost:9180", "--api-key", "test-key"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Context "local" created and saved.`)
	assert.Contains(t, out.String(), `Context "local" set as current context.`)
	assert.True(t, cfg.saveCalled)
	assert.Len(t, cfg.contexts, 1)
	assert.Equal(t, "local", cfg.contexts[0].Name)
	assert.Equal(t, "http://localhost:9180", cfg.contexts[0].Server)
	assert.Equal(t, "test-key", cfg.contexts[0].APIKey)
}

func TestCreate_SecondContext_NotAutoSetCurrent(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := &mockConfig{
		contexts:       []config.Context{{Name: "existing", Server: "http://existing:9180"}},
		currentContext: "existing",
	}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"new-ctx", "--server", "http://new:9180"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `Context "new-ctx" created and saved.`)
	assert.NotContains(t, out.String(), `set as current context`)
	assert.Equal(t, "existing", cfg.currentContext)
}

func TestCreate_Duplicate(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
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

	c := NewCmdCreate(f)
	c.SetArgs([]string{"local", "--server", "http://other:9180"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCreate_SaveError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{saveErr: fmt.Errorf("disk full")}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"local", "--server", "http://localhost:9180"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}

func TestCreate_MissingServerFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := &mockConfig{}

	f := &cmd.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	c := NewCmdCreate(f)
	c.SetArgs([]string{"local"})
	err := c.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server")
}
