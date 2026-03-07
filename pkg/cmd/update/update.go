package update

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/update"
	"github.com/api7/a6/internal/version"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type Options struct {
	IO      *iostreams.IOStreams
	Version string
	Force   bool

	currentVersion func() string

	fetchLatestRelease func() (update.Release, error)
	findAsset          func(update.Release) (update.Asset, error)
	download           func(update.Asset, io.Writer) (string, error)
	install            func(string) error
}

func NewCmdUpdate(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:             f.IOStreams,
		currentVersion: func() string { return version.Version },

		fetchLatestRelease: update.FetchLatestRelease,
		findAsset:          update.FindAsset,
		download:           update.Download,
		install:            update.Install,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a6 to the latest release",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return updateRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")

	return cmd
}

func updateRun(opts *Options) error {
	currentVersion := ""
	if opts.currentVersion != nil {
		currentVersion = strings.TrimSpace(opts.currentVersion())
	}
	fmt.Fprintf(opts.IO.Out, "Current version: %s\n", currentVersion)

	isDev := currentVersion == "dev"
	if isDev {
		fmt.Fprintln(opts.IO.ErrOut, "Warning: current build is a dev build; update comparison may be inaccurate.")
		proceed, err := confirmUpdate(opts, "Continue with update from latest release? (y/N): ")
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
	}

	release, err := opts.fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		fmt.Fprintln(opts.IO.Out, "No published releases found.")
		return nil
	}

	if !isDev {
		newer, err := update.HasNewerVersion(currentVersion, release.TagName)
		if err != nil {
			return fmt.Errorf("failed to compare versions: %w", err)
		}
		if !newer {
			fmt.Fprintf(opts.IO.Out, "a6 is already up to date (%s).\n", currentVersion)
			return nil
		}
	}

	fmt.Fprintf(opts.IO.Out, "Updating: %s -> %s\n", currentVersion, release.TagName)
	if strings.TrimSpace(release.Name) != "" {
		fmt.Fprintf(opts.IO.Out, "Release: %s\n", strings.TrimSpace(release.Name))
	}
	if summary := summarizeReleaseBody(release.Body); summary != "" {
		fmt.Fprintf(opts.IO.Out, "Notes: %s\n", summary)
	}

	if !isDev {
		proceed, err := confirmUpdate(opts, "Proceed with update? (y/N): ")
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
	}

	asset, err := opts.findAsset(release)
	if err != nil {
		return fmt.Errorf("failed to find release asset: %w", err)
	}

	if opts.IO.IsStdoutTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Downloading %s ...\n", asset.Name)
	}
	tmpFile, err := opts.download(asset, nil)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tmpFile)

	if err := opts.install(tmpFile); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	fmt.Fprintf(opts.IO.Out, "Updated successfully to %s\n", release.TagName)
	if strings.TrimSpace(release.HTMLURL) != "" {
		fmt.Fprintf(opts.IO.Out, "Release URL: %s\n", strings.TrimSpace(release.HTMLURL))
	}

	return nil
}

func confirmUpdate(opts *Options, prompt string) (bool, error) {
	if opts.Force || !opts.IO.IsStdinTTY() {
		return true, nil
	}

	fmt.Fprint(opts.IO.ErrOut, prompt)
	reader := bufio.NewReader(opts.IO.In)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Fprintln(opts.IO.ErrOut, "Aborted.")
		return false, nil
	}
	return true, nil
}

func summarizeReleaseBody(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line != "" {
			return line
		}
	}
	return ""
}
