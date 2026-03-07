package extension

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name        string `yaml:"name" json:"name"`
	Owner       string `yaml:"owner" json:"owner"`
	Repo        string `yaml:"repo" json:"repo"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
	BinaryPath  string `yaml:"path" json:"path"`
}

type Extension struct {
	Manifest
	Dir  string `json:"dir" yaml:"dir"`
	Path string `json:"binary" yaml:"binary"`
}

type Manager struct {
	extensionsDir string
	fetchRelease  func(owner, repo string) (Release, error)
	downloadAsset func(url string, w io.Writer) (string, error)
}

func NewManager(extensionsDir string) *Manager {
	return &Manager{
		extensionsDir: extensionsDir,
		fetchRelease:  FetchLatestRelease,
		downloadAsset: downloadAsset,
	}
}

func DefaultExtensionsDir() string {
	if dir := os.Getenv("A6_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "extensions")
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "a6", "extensions")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "a6", "extensions")
}

func (m *Manager) Install(ownerRepo string) (*Extension, error) {
	owner, repo, err := parseOwnerRepo(ownerRepo)
	if err != nil {
		return nil, err
	}
	name, err := extensionNameFromRepo(repo)
	if err != nil {
		return nil, err
	}

	release, err := m.fetchRelease(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, fmt.Errorf("no published releases found for %s/%s", owner, repo)
	}

	asset, err := findAsset(release, name)
	if err != nil {
		return nil, fmt.Errorf("find release asset: %w", err)
	}

	downloadedPath, err := m.downloadAsset(asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("download release asset: %w", err)
	}
	defer os.Remove(downloadedPath)

	extDir := filepath.Join(m.extensionsDir, "a6-"+name)
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("create extension directory: %w", err)
	}

	binaryName := extensionBinaryName(name)
	binaryAbs := filepath.Join(extDir, binaryName)
	if err := extractExtensionBinary(downloadedPath, binaryAbs, binaryName); err != nil {
		return nil, err
	}

	manifest := Manifest{
		Name:        name,
		Owner:       owner,
		Repo:        repo,
		Version:     normalizeVersion(release.TagName),
		Description: firstNonEmpty(strings.TrimSpace(release.Name), summarizeBody(release.Body)),
		BinaryPath:  binaryName,
	}
	if err := writeManifest(filepath.Join(extDir, "manifest.yaml"), manifest); err != nil {
		return nil, err
	}

	return &Extension{Manifest: manifest, Dir: extDir, Path: binaryAbs}, nil
}

func (m *Manager) List() ([]Extension, error) {
	if _, err := os.Stat(m.extensionsDir); err != nil {
		if os.IsNotExist(err) {
			return []Extension{}, nil
		}
		return nil, fmt.Errorf("stat extensions directory: %w", err)
	}

	entries, err := os.ReadDir(m.extensionsDir)
	if err != nil {
		return nil, fmt.Errorf("read extensions directory: %w", err)
	}

	exts := make([]Extension, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ext, err := readExtensionFromDir(filepath.Join(m.extensionsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		exts = append(exts, ext)
	}

	sort.Slice(exts, func(i, j int) bool { return exts[i].Name < exts[j].Name })
	return exts, nil
}

func (m *Manager) Find(name string) (*Extension, error) {
	exts, err := m.List()
	if err != nil {
		return nil, err
	}
	for i := range exts {
		if exts[i].Name == name {
			ext := exts[i]
			return &ext, nil
		}
	}
	return nil, fmt.Errorf("extension %q not found", name)
}

func (m *Manager) Upgrade(name string) (*Extension, error) {
	current, err := m.Find(name)
	if err != nil {
		return nil, err
	}
	updated, _, err := m.upgradeInstalled(*current)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return current, nil
	}
	return updated, nil
}

func (m *Manager) UpgradeAll() ([]Extension, error) {
	exts, err := m.List()
	if err != nil {
		return nil, err
	}

	upgraded := make([]Extension, 0)
	for i := range exts {
		ext := exts[i]
		updated, changed, err := m.upgradeInstalled(ext)
		if err != nil {
			return nil, err
		}
		if changed && updated != nil {
			upgraded = append(upgraded, *updated)
		}
	}

	return upgraded, nil
}

func (m *Manager) Remove(name string) error {
	ext, err := m.Find(name)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(ext.Dir); err != nil {
		return fmt.Errorf("remove extension %q: %w", name, err)
	}
	return nil
}

func (m *Manager) upgradeInstalled(ext Extension) (*Extension, bool, error) {
	release, err := m.fetchRelease(ext.Owner, ext.Repo)
	if err != nil {
		return nil, false, fmt.Errorf("fetch latest release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, false, fmt.Errorf("no published releases found for %s/%s", ext.Owner, ext.Repo)
	}

	newVersion := normalizeVersion(release.TagName)
	if !isVersionNewer(ext.Version, newVersion) {
		return &ext, false, nil
	}

	asset, err := findAsset(release, ext.Name)
	if err != nil {
		return nil, false, fmt.Errorf("find release asset: %w", err)
	}

	downloadedPath, err := m.downloadAsset(asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("download release asset: %w", err)
	}
	defer os.Remove(downloadedPath)

	binaryName := ext.BinaryPath
	if strings.TrimSpace(binaryName) == "" {
		binaryName = extensionBinaryName(ext.Name)
	}
	binaryAbs := filepath.Join(ext.Dir, binaryName)
	if err := extractExtensionBinary(downloadedPath, binaryAbs, filepath.Base(binaryName)); err != nil {
		return nil, false, err
	}

	ext.Version = newVersion
	ext.Description = firstNonEmpty(strings.TrimSpace(release.Name), summarizeBody(release.Body), ext.Description)
	ext.BinaryPath = binaryName
	if err := writeManifest(filepath.Join(ext.Dir, "manifest.yaml"), ext.Manifest); err != nil {
		return nil, false, err
	}
	ext.Path = binaryAbs

	return &ext, true, nil
}

func readExtensionFromDir(dir string) (Extension, error) {
	manifestPath := filepath.Join(dir, "manifest.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return Extension{}, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return Extension{}, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return Extension{}, fmt.Errorf("invalid manifest %s: name is empty", manifestPath)
	}
	if strings.TrimSpace(manifest.BinaryPath) == "" {
		manifest.BinaryPath = extensionBinaryName(manifest.Name)
	}

	return Extension{Manifest: manifest, Dir: dir, Path: filepath.Join(dir, manifest.BinaryPath)}, nil
}

func writeManifest(path string, manifest Manifest) error {
	data, err := yaml.Marshal(&manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

func parseOwnerRepo(ownerRepo string) (string, string, error) {
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("owner/repo must be in the format <owner>/<repo>")
	}
	return parts[0], parts[1], nil
}

func extensionNameFromRepo(repo string) (string, error) {
	if !strings.HasPrefix(repo, "a6-") || len(repo) <= len("a6-") {
		return "", fmt.Errorf("extension repository must be named a6-<name>")
	}
	return strings.TrimPrefix(repo, "a6-"), nil
}

func extensionBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return "a6-" + name + ".exe"
	}
	return "a6-" + name
}

func extractExtensionBinary(downloadedPath, destinationPath, binaryName string) error {
	format, err := detectArchive(downloadedPath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create extension binary directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".a6-ext-*")
	if err != nil {
		return fmt.Errorf("create temp extension file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	switch format {
	case "tar.gz":
		err = extractFromTarGz(downloadedPath, binaryName, tmp)
	case "zip":
		err = extractFromZip(downloadedPath, binaryName, tmp)
	default:
		err = copyFileToWriter(downloadedPath, tmp)
	}
	if err != nil {
		cleanup()
		return err
	}

	if runtime.GOOS != "windows" {
		if err := tmp.Chmod(0o755); err != nil {
			cleanup()
			return fmt.Errorf("set executable permissions: %w", err)
		}
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp extension file: %w", err)
	}

	if err := os.Rename(tmpPath, destinationPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace extension binary: %w", err)
	}

	return nil
}

func copyFileToWriter(path string, dst io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open downloaded file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(dst, f); err != nil {
		return fmt.Errorf("copy downloaded binary: %w", err)
	}
	return nil
}

func detectArchive(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open downloaded file: %w", err)
	}
	defer f.Close()

	head := make([]byte, 4)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("read downloaded file header: %w", err)
	}
	head = head[:n]

	if len(head) >= 2 && head[0] == 0x1f && head[1] == 0x8b {
		return "tar.gz", nil
	}
	if len(head) >= 4 && head[0] == 'P' && head[1] == 'K' && head[2] == 0x03 && head[3] == 0x04 {
		return "zip", nil
	}

	return "binary", nil
}

func extractFromTarGz(path, binaryName string, dst io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("open gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive entry: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != binaryName {
			continue
		}
		if _, err := io.Copy(dst, tr); err != nil {
			return fmt.Errorf("extract binary: %w", err)
		}
		return nil
	}

	return fmt.Errorf("binary %q not found in archive", binaryName)
}

func extractFromZip(path, binaryName string, dst io.Writer) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if filepath.Base(f.Name) != binaryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry: %w", err)
		}
		_, copyErr := io.Copy(dst, rc)
		closeErr := rc.Close()
		if copyErr != nil {
			return fmt.Errorf("extract binary: %w", copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close zip entry: %w", closeErr)
		}
		return nil
	}

	return fmt.Errorf("binary %q not found in archive", binaryName)
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func summarizeBody(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	for _, line := range strings.Split(trimmed, "\n") {
		candidate := strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func isVersionNewer(current, candidate string) bool {
	a := parseVersionParts(current)
	b := parseVersionParts(candidate)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		av := 0
		bv := 0
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if bv > av {
			return true
		}
		if bv < av {
			return false
		}
	}
	return false
}

func parseVersionParts(v string) []int {
	normalized := normalizeVersion(v)
	parts := strings.Split(normalized, ".")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := part
		if idx := strings.Index(trimmed, "-"); idx >= 0 {
			trimmed = trimmed[:idx]
		}
		n, err := strconv.Atoi(trimmed)
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}
