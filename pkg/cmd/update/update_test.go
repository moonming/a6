package update

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/internal/update"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type mockConfig struct{}

func (m *mockConfig) BaseURL() string                                 { return "" }
func (m *mockConfig) APIKey() string                                  { return "" }
func (m *mockConfig) CurrentContext() string                          { return "" }
func (m *mockConfig) Contexts() []config.Context                      { return nil }
func (m *mockConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (m *mockConfig) AddContext(ctx config.Context) error             { return nil }
func (m *mockConfig) RemoveContext(name string) error                 { return nil }
func (m *mockConfig) SetCurrentContext(name string) error             { return nil }
func (m *mockConfig) Save() error                                     { return nil }

func TestUpdateCommand_HelpShowsForceFlag(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	f := &cmd.Factory{
		IOStreams:  ios,
		HttpClient: func() (*http.Client, error) { return nil, nil },
		Config:     func() (config.Config, error) { return &mockConfig{}, nil },
	}

	c := NewCmdUpdate(f)
	c.SetOut(out)
	c.SetErr(errOut)
	c.SetArgs([]string{"--help"})
	err := c.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "--force")
}

func TestUpdateRun_AlreadyUpToDate(t *testing.T) {
	ios, in, out, errOut := iostreams.Test()
	_ = in

	err := updateRun(&Options{
		IO:             ios,
		currentVersion: func() string { return "v1.2.3" },
		fetchLatestRelease: func() (update.Release, error) {
			return update.Release{TagName: "v1.2.3"}, nil
		},
		findAsset: func(update.Release) (update.Asset, error) {
			return update.Asset{}, errors.New("should not be called")
		},
		download: func(update.Asset, io.Writer) (string, error) {
			return "", errors.New("should not be called")
		},
		install: func(string) error {
			return errors.New("should not be called")
		},
	})
	require.NoError(t, err)
	assert.Contains(t, out.String(), "already up to date")
	assert.Empty(t, errOut.String())
}

func TestUpdateRun_SuccessWithForce(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	called := false

	err := updateRun(&Options{
		IO:             ios,
		currentVersion: func() string { return "v1.0.0" },
		Force:          true,
		fetchLatestRelease: func() (update.Release, error) {
			return update.Release{
				TagName: "v1.1.0",
				Name:    "v1.1.0",
				Body:    "- Bug fixes\n- Improvements",
				HTMLURL: "https://github.com/api7/a6/releases/tag/v1.1.0",
			}, nil
		},
		findAsset: func(r update.Release) (update.Asset, error) {
			assert.Equal(t, "v1.1.0", r.TagName)
			return update.Asset{Name: "a6_1.1.0_darwin_arm64.tar.gz"}, nil
		},
		download: func(a update.Asset, w io.Writer) (string, error) {
			assert.Equal(t, "a6_1.1.0_darwin_arm64.tar.gz", a.Name)
			assert.Nil(t, w)
			return "/tmp/a6.bin", nil
		},
		install: func(path string) error {
			called = true
			assert.Equal(t, "/tmp/a6.bin", path)
			return nil
		},
	})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Contains(t, out.String(), "Updating: v1.0.0 -> v1.1.0")
	assert.Contains(t, out.String(), "Updated successfully to v1.1.0")
}

func TestUpdateRun_DevBuildGraceful(t *testing.T) {
	ios, in, out, errOut := iostreams.Test()
	ios.SetStdinTTY(true)
	_, _ = fmt.Fprintln(in, "y")

	err := updateRun(&Options{
		IO:             ios,
		currentVersion: func() string { return "dev" },
		Force:          false,
		fetchLatestRelease: func() (update.Release, error) {
			return update.Release{TagName: "v1.2.0"}, nil
		},
		findAsset: func(update.Release) (update.Asset, error) {
			return update.Asset{Name: "a6_1.2.0_linux_amd64.tar.gz"}, nil
		},
		download: func(update.Asset, io.Writer) (string, error) {
			return "/tmp/a6.bin", nil
		},
		install: func(string) error {
			return nil
		},
	})
	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "dev build")
	assert.Contains(t, out.String(), "Updated successfully")
}

func TestUpdateRun_ConfirmAbort(t *testing.T) {
	ios, in, out, errOut := iostreams.Test()
	ios.SetStdinTTY(true)
	_, _ = fmt.Fprintln(in, "n")

	err := updateRun(&Options{
		IO:             ios,
		currentVersion: func() string { return "v1.0.0" },
		fetchLatestRelease: func() (update.Release, error) {
			return update.Release{TagName: "v1.1.0"}, nil
		},
		findAsset: func(update.Release) (update.Asset, error) {
			return update.Asset{}, errors.New("must not be called")
		},
		download: func(update.Asset, io.Writer) (string, error) {
			return "", errors.New("must not be called")
		},
		install: func(string) error {
			return errors.New("must not be called")
		},
	})
	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "Aborted.")
	assert.Contains(t, out.String(), "Updating:")
}

func TestSummarizeReleaseBody(t *testing.T) {
	assert.Equal(t, "Feature A", summarizeReleaseBody("- Feature A\n- Feature B"))
	assert.Equal(t, "plain line", summarizeReleaseBody("plain line\n- next"))
	assert.Equal(t, "", summarizeReleaseBody("\n\n"))
}

func TestConfirmUpdate_NonTTYSkipsPrompt(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ok, err := confirmUpdate(&Options{IO: ios}, "Proceed? ")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestUpdateRun_FetchReleaseError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := updateRun(&Options{
		IO:             ios,
		currentVersion: func() string { return "v1.0.0" },
		fetchLatestRelease: func() (update.Release, error) {
			return update.Release{}, errors.New("boom")
		},
	})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "fetch latest release") || strings.Contains(err.Error(), "failed"))
}
