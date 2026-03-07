package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/internal/update"
	ver "github.com/api7/a6/internal/version"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/root"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

func main() {
	ios := iostreams.System()

	cfg := config.NewFileConfig()

	// Apply environment variable overrides.
	if v := os.Getenv("A6_SERVER"); v != "" {
		cfg.SetServerOverride(v)
	}
	if v := os.Getenv("A6_API_KEY"); v != "" {
		cfg.SetAPIKeyOverride(v)
	}

	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			apiKey := cfg.APIKey()
			if apiKey == "" {
				return nil, fmt.Errorf("no API key configured; use 'a6 context create' or set A6_API_KEY")
			}
			return api.NewAuthenticatedClient(apiKey), nil
		},
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}

	rootCmd := root.NewCmdRoot(f)

	// Wire flag overrides into config after flags are parsed.
	rootCmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		if v, _ := c.Flags().GetString("server"); v != "" {
			cfg.SetServerOverride(v)
		}
		if v, _ := c.Flags().GetString("api-key"); v != "" {
			cfg.SetAPIKeyOverride(v)
		}
		return nil
	}

	err := rootCmd.Execute()
	maybeCheckForUpdate(rootCmd, ios, os.Args[1:])

	if err != nil {
		if !cmdutil.IsSilent(err) {
			fmt.Fprintln(ios.ErrOut, err)
		}
		os.Exit(1)
	}
}

func maybeCheckForUpdate(rootCmd *cobra.Command, ios *iostreams.IOStreams, args []string) {
	if shouldSkipUpdateCheck(rootCmd, ios, args) {
		return
	}

	state, err := update.ReadState()
	if err != nil {
		return
	}

	if latestVersion, _, ok := update.UpdateAvailableFromState(state, ver.Version); ok {
		fmt.Fprintf(ios.ErrOut, "A new version of a6 is available: %s → %s\nRun 'a6 update' to update.\n", ver.Version, latestVersion)
	}

	if !update.ShouldCheck(state, time.Now()) {
		return
	}

	go func(now time.Time) {
		latestVersion, latestURL, ok := update.CheckForUpdate()
		newState := update.StateFile{CheckedAt: now}
		if ok {
			newState.LatestVersion = latestVersion
			newState.LatestURL = latestURL
		}
		_ = update.WriteState(newState)
	}(time.Now().UTC())
}

func shouldSkipUpdateCheck(rootCmd *cobra.Command, ios *iostreams.IOStreams, args []string) bool {
	if !ios.IsStdoutTTY() {
		return true
	}
	if os.Getenv("A6_NO_UPDATE_CHECK") != "" {
		return true
	}
	if os.Getenv("CI") != "" {
		return true
	}

	found, _, findErr := rootCmd.Find(args)
	if findErr != nil || found == nil {
		return false
	}
	name := found.Name()
	return name == "update" || name == "version" || name == "completion"
}
