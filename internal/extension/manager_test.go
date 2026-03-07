package extension

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestManagerInstallValidOwnerRepo(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(t.TempDir(), "a6-hello")
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755))

	mgr := NewManager(dir)
	mgr.fetchRelease = func(owner, repo string) (Release, error) {
		require.Equal(t, "api7", owner)
		require.Equal(t, "a6-hello", repo)
		return releaseFor("hello", "v1.0.0"), nil
	}
	mgr.downloadAsset = func(url string, w io.Writer) (string, error) {
		return binaryPath, nil
	}

	ext, err := mgr.Install("api7/a6-hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", ext.Name)
	assert.Equal(t, "1.0.0", ext.Version)
	assert.FileExists(t, ext.Path)
	assert.FileExists(t, filepath.Join(ext.Dir, "manifest.yaml"))
}

func TestManagerInstallInvalidFormat(t *testing.T) {
	mgr := NewManager(t.TempDir())
	_, err := mgr.Install("bad-format")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner/repo")
}

func TestManagerListEmpty(t *testing.T) {
	mgr := NewManager(filepath.Join(t.TempDir(), "extensions"))
	exts, err := mgr.List()
	require.NoError(t, err)
	assert.Empty(t, exts)
}

func TestManagerListWithInstalledExtensions(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-hello"), Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.0.0", BinaryPath: extensionBinaryName("hello")}))
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-world"), Manifest{Name: "world", Owner: "api7", Repo: "a6-world", Version: "2.0.0", BinaryPath: extensionBinaryName("world")}))

	mgr := NewManager(dir)
	exts, err := mgr.List()
	require.NoError(t, err)
	require.Len(t, exts, 2)
	assert.Equal(t, "hello", exts[0].Name)
	assert.Equal(t, "world", exts[1].Name)
}

func TestManagerFindExisting(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-hello"), Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.0.0", BinaryPath: extensionBinaryName("hello")}))

	mgr := NewManager(dir)
	ext, err := mgr.Find("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", ext.Name)
}

func TestManagerFindNonExistent(t *testing.T) {
	mgr := NewManager(t.TempDir())
	_, err := mgr.Find("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManagerRemoveExisting(t *testing.T) {
	dir := t.TempDir()
	extDir := filepath.Join(dir, "a6-hello")
	require.NoError(t, writeManifestFile(t, extDir, Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.0.0", BinaryPath: extensionBinaryName("hello")}))

	mgr := NewManager(dir)
	require.NoError(t, mgr.Remove("hello"))
	_, err := os.Stat(extDir)
	assert.True(t, os.IsNotExist(err))
}

func TestManagerRemoveNonExistent(t *testing.T) {
	mgr := NewManager(t.TempDir())
	err := mgr.Remove("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManagerUpgradeNewerVersion(t *testing.T) {
	dir := t.TempDir()
	extDir := filepath.Join(dir, "a6-hello")
	require.NoError(t, writeManifestFile(t, extDir, Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.0.0", BinaryPath: extensionBinaryName("hello")}))

	newBinaryPath := filepath.Join(t.TempDir(), "new-a6-hello")
	require.NoError(t, os.WriteFile(newBinaryPath, []byte("new"), 0o755))

	mgr := NewManager(dir)
	mgr.fetchRelease = func(owner, repo string) (Release, error) {
		return releaseFor("hello", "v1.1.0"), nil
	}
	mgr.downloadAsset = func(url string, w io.Writer) (string, error) {
		return newBinaryPath, nil
	}

	ext, err := mgr.Upgrade("hello")
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", ext.Version)
}

func TestManagerUpgradeAlreadyLatest(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-hello"), Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.1.0", BinaryPath: extensionBinaryName("hello")}))

	mgr := NewManager(dir)
	mgr.fetchRelease = func(owner, repo string) (Release, error) {
		return releaseFor("hello", "v1.1.0"), nil
	}
	mgr.downloadAsset = func(url string, w io.Writer) (string, error) {
		return "", assert.AnError
	}

	ext, err := mgr.Upgrade("hello")
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", ext.Version)
}

func TestManagerUpgradeAll(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-hello"), Manifest{Name: "hello", Owner: "api7", Repo: "a6-hello", Version: "1.0.0", BinaryPath: extensionBinaryName("hello")}))
	require.NoError(t, writeManifestFile(t, filepath.Join(dir, "a6-world"), Manifest{Name: "world", Owner: "api7", Repo: "a6-world", Version: "2.0.0", BinaryPath: extensionBinaryName("world")}))

	binary := filepath.Join(t.TempDir(), "a6-binary")
	require.NoError(t, os.WriteFile(binary, []byte("bin"), 0o755))

	mgr := NewManager(dir)
	mgr.fetchRelease = func(owner, repo string) (Release, error) {
		switch repo {
		case "a6-hello":
			return releaseFor("hello", "v1.1.0"), nil
		case "a6-world":
			return releaseFor("world", "v2.0.0"), nil
		default:
			return Release{}, nil
		}
	}
	mgr.downloadAsset = func(url string, w io.Writer) (string, error) {
		return binary, nil
	}

	upgraded, err := mgr.UpgradeAll()
	require.NoError(t, err)
	require.Len(t, upgraded, 1)
	assert.Equal(t, "hello", upgraded[0].Name)
	assert.Equal(t, "1.1.0", upgraded[0].Version)
}

func releaseFor(name, tag string) Release {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return Release{
		TagName: tag,
		Name:    tag,
		Assets: []Asset{{
			Name:               "a6-" + name + "_" + normalizeVersion(tag) + "_" + runtime.GOOS + "_" + runtime.GOARCH + ext,
			BrowserDownloadURL: "https://example.com/download",
		}},
	}
}

func writeManifestFile(t *testing.T, dir string, manifest Manifest) error {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	binaryPath := filepath.Join(dir, manifest.BinaryPath)
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(&manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.yaml"), b, 0o644)
}
