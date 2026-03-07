package extension

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const githubAPIBaseURL = "https://api.github.com"

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func fetchLatestRelease(baseURL string, client *http.Client, owner, repo string) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(baseURL, "/"), owner, repo)

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

func FetchLatestRelease(owner, repo string) (Release, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	return fetchLatestRelease(githubAPIBaseURL, client, owner, repo)
}

func findAsset(release Release, name string) (Asset, error) {
	if strings.TrimSpace(release.TagName) == "" {
		return Asset{}, fmt.Errorf("release tag is empty")
	}

	version := normalizeVersion(release.TagName)
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	target := fmt.Sprintf("a6-%s_%s_%s_%s%s", name, version, runtime.GOOS, runtime.GOARCH, ext)
	targetWithV := fmt.Sprintf("a6-%s_v%s_%s_%s%s", name, version, runtime.GOOS, runtime.GOARCH, ext)

	for _, asset := range release.Assets {
		if asset.Name == target || asset.Name == targetWithV {
			return asset, nil
		}
	}

	return Asset{}, fmt.Errorf("no matching release asset for extension %q on %s/%s", name, runtime.GOOS, runtime.GOARCH)
}

func downloadAsset(url string, w io.Writer) (string, error) {
	if strings.TrimSpace(url) == "" {
		return "", fmt.Errorf("asset download URL is empty")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download release asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "a6-extension-download-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	var dst io.Writer = tmp
	if w != nil {
		dst = io.MultiWriter(tmp, w)
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write downloaded asset: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("close downloaded asset: %w", err)
	}

	return tmpPath, nil
}
