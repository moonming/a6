package extension

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchLatestReleaseSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/repos/acme/a6-hello/releases/latest", r.URL.Path)
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","name":"v1.2.3","body":"notes","assets":[{"name":"a6-hello_1.2.3_linux_amd64.tar.gz","browser_download_url":"https://example.com/a6-hello.tar.gz","size":123}]}`))
	}))
	defer srv.Close()

	release, err := fetchLatestRelease(srv.URL, &http.Client{Timeout: 2 * time.Second}, "acme", "a6-hello")
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3", release.TagName)
	assert.Len(t, release.Assets, 1)
}

func TestFetchLatestReleaseNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	release, err := fetchLatestRelease(srv.URL, &http.Client{Timeout: 2 * time.Second}, "acme", "a6-hello")
	require.NoError(t, err)
	assert.Equal(t, Release{}, release)
}

func TestFindAssetCurrentPlatform(t *testing.T) {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	release := Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "a6-hello_1.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ext},
			{Name: "a6-hello_1.0.0_linux_386.tar.gz"},
		},
	}

	asset, err := findAsset(release, "hello")
	require.NoError(t, err)
	assert.Equal(t, "a6-hello_1.0.0_"+runtime.GOOS+"_"+runtime.GOARCH+ext, asset.Name)
}

func TestFindAssetNoMatch(t *testing.T) {
	release := Release{
		TagName: "v1.0.0",
		Assets:  []Asset{{Name: "a6-hello_1.0.0_linux_386.tar.gz"}},
	}

	_, err := findAsset(release, "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no matching release asset")
}
