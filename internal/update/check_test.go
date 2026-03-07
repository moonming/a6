package update

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateFilePath_UsesA6ConfigDir(t *testing.T) {
	t.Setenv("A6_CONFIG_DIR", "/tmp/a6-cfg")
	t.Setenv("XDG_CONFIG_HOME", "")
	assert.Equal(t, "/tmp/a6-cfg/update-check.json", StateFilePath())
}

func TestStateFilePath_UsesXDGConfigHome(t *testing.T) {
	t.Setenv("A6_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	assert.Equal(t, "/tmp/xdg/a6/update-check.json", StateFilePath())
}

func TestShouldCheck(t *testing.T) {
	now := time.Now()
	assert.True(t, ShouldCheck(StateFile{}, now))
	assert.False(t, ShouldCheck(StateFile{CheckedAt: now.Add(-23 * time.Hour)}, now))
	assert.True(t, ShouldCheck(StateFile{CheckedAt: now.Add(-24 * time.Hour)}, now))
	assert.True(t, ShouldCheck(StateFile{CheckedAt: now.Add(-25 * time.Hour)}, now))
}

func TestReadWriteState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("A6_CONFIG_DIR", dir)
	t.Setenv("XDG_CONFIG_HOME", "")

	state := StateFile{
		CheckedAt:     time.Now().UTC().Truncate(time.Second),
		LatestVersion: "v1.2.3",
		LatestURL:     "https://github.com/api7/a6/releases/tag/v1.2.3",
	}
	require.NoError(t, WriteState(state))

	got, err := ReadState()
	require.NoError(t, err)
	assert.Equal(t, state.CheckedAt, got.CheckedAt)
	assert.Equal(t, state.LatestVersion, got.LatestVersion)
	assert.Equal(t, state.LatestURL, got.LatestURL)

	path := filepath.Join(dir, "update-check.json")
	assert.FileExists(t, path)
}

func TestReadState_NotExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("A6_CONFIG_DIR", dir)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := ReadState()
	require.NoError(t, err)
	assert.Equal(t, StateFile{}, got)
}

func TestUpdateAvailableFromState(t *testing.T) {
	state := StateFile{LatestVersion: "v1.3.0", LatestURL: "https://example.com"}
	v, u, ok := UpdateAvailableFromState(state, "v1.2.0")
	assert.True(t, ok)
	assert.Equal(t, "v1.3.0", v)
	assert.Equal(t, "https://example.com", u)

	v, u, ok = UpdateAvailableFromState(state, "v1.3.0")
	assert.False(t, ok)
	assert.Empty(t, v)
	assert.Empty(t, u)
}

func TestHasNewerVersion(t *testing.T) {
	ok, err := HasNewerVersion("dev", "v1.0.0")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = HasNewerVersion("v1.2.0", "v1.1.0")
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = HasNewerVersion("invalid", "v1.1.0")
	require.Error(t, err)
}
