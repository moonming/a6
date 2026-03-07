package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/api7/a6/internal/version"
)

type StateFile struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	LatestURL     string    `json:"latest_url"`
}

const CheckInterval = 24 * time.Hour

func StateFilePath() string {
	if dir := os.Getenv("A6_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "update-check.json")
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "a6", "update-check.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "a6", "update-check.json")
}

func ShouldCheck(state StateFile, now time.Time) bool {
	if state.CheckedAt.IsZero() {
		return true
	}
	return now.Sub(state.CheckedAt) >= CheckInterval
}

func ReadState() (StateFile, error) {
	b, err := os.ReadFile(StateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return StateFile{}, nil
		}
		return StateFile{}, err
	}

	var state StateFile
	if err := json.Unmarshal(b, &state); err != nil {
		return StateFile{}, err
	}
	return state, nil
}

func WriteState(state StateFile) error {
	path := StateFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func HasNewerVersion(currentVersion, latestVersion string) (bool, error) {
	current, err := ParseSemver(strings.TrimSpace(currentVersion))
	if err != nil {
		return false, err
	}
	latest, err := ParseSemver(strings.TrimSpace(latestVersion))
	if err != nil {
		return false, err
	}
	return current.IsNewer(latest), nil
}

func UpdateAvailableFromState(state StateFile, currentVersion string) (string, string, bool) {
	if strings.TrimSpace(state.LatestVersion) == "" {
		return "", "", false
	}
	newer, err := HasNewerVersion(currentVersion, state.LatestVersion)
	if err != nil {
		return "", "", false
	}
	if !newer {
		return "", "", false
	}
	return state.LatestVersion, state.LatestURL, true
}

func CheckForUpdate() (string, string, bool) {
	release, err := FetchLatestRelease()
	if err != nil {
		return "", "", false
	}
	if strings.TrimSpace(release.TagName) == "" {
		return "", "", false
	}

	newer, err := HasNewerVersion(version.Version, release.TagName)
	if err != nil {
		return "", "", false
	}
	if !newer {
		return "", "", false
	}

	return release.TagName, release.HTMLURL, true
}
