package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	GitHubOwner = "api7"
	GitHubRepo  = "a6"
)

const githubAPIBaseURL = "https://api.github.com"

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

func fetchLatestRelease(baseURL string, client *http.Client) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(baseURL, "/"), GitHubOwner, GitHubRepo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, fmt.Errorf("build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("request latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return Release{}, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decode latest release: %w", err)
	}

	return release, nil
}

func FetchLatestRelease() (Release, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return fetchLatestRelease(githubAPIBaseURL, client)
}

func FindAsset(release Release) (Asset, error) {
	if strings.TrimSpace(release.TagName) == "" {
		return Asset{}, fmt.Errorf("release tag is empty")
	}

	version := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	target := fmt.Sprintf("a6_%s_%s_%s%s", version, runtime.GOOS, runtime.GOARCH, ext)

	for _, asset := range release.Assets {
		if asset.Name == target {
			return asset, nil
		}
	}

	return Asset{}, fmt.Errorf("no matching release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
}
