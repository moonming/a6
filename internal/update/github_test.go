package update

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchLatestRelease_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/repos/api7/a6/releases/latest", r.URL.Path)
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v1.2.3",
			"name": "v1.2.3",
			"body": "release notes",
			"html_url": "https://github.com/api7/a6/releases/tag/v1.2.3",
			"assets": [{
				"name": "a6_1.2.3_linux_amd64.tar.gz",
				"browser_download_url": "https://example.com/a6.tar.gz",
				"size": 123,
				"content_type": "application/gzip"
			}]
		}`))
	}))
	defer srv.Close()

	release, err := fetchLatestRelease(srv.URL, &http.Client{Timeout: 2 * time.Second})
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", release.TagName)
	assert.Equal(t, "v1.2.3", release.Name)
	assert.Len(t, release.Assets, 1)
	assert.Equal(t, "a6_1.2.3_linux_amd64.tar.gz", release.Assets[0].Name)
}

func TestFetchLatestRelease_404ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	release, err := fetchLatestRelease(srv.URL, &http.Client{Timeout: 2 * time.Second})
	require.NoError(t, err)
	assert.Equal(t, Release{}, release)
}

func TestFetchLatestRelease_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := fetchLatestRelease(srv.URL, &http.Client{Timeout: 2 * time.Second})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestFindAsset_MatchCurrentPlatform(t *testing.T) {
	release := Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "a6_1.0.0_linux_amd64.tar.gz"},
			{Name: "a6_1.0.0_darwin_arm64.tar.gz"},
			{Name: "a6_1.0.0_windows_amd64.zip"},
		},
	}

	asset, err := FindAsset(release)
	require.NoError(t, err)

	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	assert.Equal(t, "a6_1.0.0_"+runtime.GOOS+"_"+runtime.GOARCH+ext, asset.Name)
}

func TestFindAsset_NoMatch(t *testing.T) {
	release := Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "a6_1.0.0_linux_386.tar.gz"},
		},
	}

	_, err := FindAsset(release)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no matching release asset")
}
