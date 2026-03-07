package upgrade

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

type mockUpgradeManager struct {
	findExt     *extension.Extension
	upgradeExt  *extension.Extension
	upgradeAll  []extension.Extension
	findErr     error
	upgradeErr  error
	upgradeAErr error
}

func (m *mockUpgradeManager) Find(name string) (*extension.Extension, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.findExt, nil
}

func (m *mockUpgradeManager) Upgrade(name string) (*extension.Extension, error) {
	if m.upgradeErr != nil {
		return nil, m.upgradeErr
	}
	return m.upgradeExt, nil
}

func (m *mockUpgradeManager) UpgradeAll() ([]extension.Extension, error) {
	if m.upgradeAErr != nil {
		return nil, m.upgradeAErr
	}
	return m.upgradeAll, nil
}

func TestUpgradeRunUpgraded(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockUpgradeManager{findExt: &extension.Extension{Manifest: extension.Manifest{Name: "hello", Version: "1.0.0"}}, upgradeExt: &extension.Extension{Manifest: extension.Manifest{Name: "hello", Version: "1.1.0"}}}}
	err := upgradeRun(opts, "hello")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Upgraded hello from v1.0.0 to v1.1.0")
}

func TestUpgradeRunAlreadyLatest(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockUpgradeManager{findExt: &extension.Extension{Manifest: extension.Manifest{Name: "hello", Version: "1.1.0"}}, upgradeExt: &extension.Extension{Manifest: extension.Manifest{Name: "hello", Version: "1.1.0"}}}}
	err := upgradeRun(opts, "hello")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "already up to date")
}

func TestUpgradeAllRun(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockUpgradeManager{upgradeAll: []extension.Extension{{Manifest: extension.Manifest{Name: "hello", Version: "1.1.0"}}}}}
	err := upgradeAllRun(opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Upgraded hello to v1.1.0")
}

func TestUpgradeAllRunNone(t *testing.T) {
	io, _, out, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockUpgradeManager{upgradeAll: []extension.Extension{}}}
	err := upgradeAllRun(opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "already up to date")
}

func TestUpgradeCommandMutualExclusive(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: io, HttpClient: func() (*http.Client, error) { return nil, nil }}
	c := NewCmdUpgrade(f)
	c.SetArgs([]string{"hello", "--all"})
	err := c.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestUpgradeRunError(t *testing.T) {
	io, _, _, _ := iostreams.Test()
	opts := &Options{IO: io, Manager: &mockUpgradeManager{findErr: fmt.Errorf("missing")}}
	err := upgradeRun(opts, "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}
