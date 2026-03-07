package install

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type mockInstallManager struct {
	ext *extension.Extension
	err error
}

func (m *mockInstallManager) Install(ownerRepo string) (*extension.Extension, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ext, nil
}

func TestInstallRunSuccess(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{
		IO:      io,
		Manager: &mockInstallManager{ext: &extension.Extension{Manifest: extension.Manifest{Name: "hello", Version: "1.2.3"}}},
	}

	err := installRun(opts, "api7/a6-hello")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Installed extension hello v1.2.3")
}

func TestInstallRunError(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	opts := &Options{
		IO:      io,
		Manager: &mockInstallManager{err: fmt.Errorf("boom")},
	}

	err := installRun(opts, "api7/a6-hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestInstallCommandInvalidFormat(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: io, HttpClient: func() (*http.Client, error) { return nil, nil }}
	c := NewCmdInstall(f)
	c.SetArgs([]string{"bad"})
	err := c.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner/repo")
}
