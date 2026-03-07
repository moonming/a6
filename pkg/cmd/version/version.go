package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	ver "github.com/api7/a6/internal/version"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
)

// Info holds the version information for display.
type Info struct {
	Version   string `json:"version" yaml:"version"`
	Commit    string `json:"commit" yaml:"commit"`
	BuildDate string `json:"buildDate" yaml:"buildDate"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	Platform  string `json:"platform" yaml:"platform"`
}

// Options holds the command options.
type Options struct {
	IO     *iostreams.IOStreams
	Output string
}

// NewCmdVersion creates the version command.
func NewCmdVersion(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return versionRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml")

	return cmd
}

func versionRun(opts *Options) error {
	info := Info{
		Version:   ver.Version,
		Commit:    ver.Commit,
		BuildDate: ver.Date,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	format := opts.Output
	if format == "" {
		fmt.Fprintf(opts.IO.Out, "a6 version %s\n", info.Version)
		fmt.Fprintf(opts.IO.Out, "  commit:    %s\n", info.Commit)
		fmt.Fprintf(opts.IO.Out, "  built:     %s\n", info.BuildDate)
		fmt.Fprintf(opts.IO.Out, "  go:        %s\n", info.GoVersion)
		fmt.Fprintf(opts.IO.Out, "  platform:  %s\n", info.Platform)
		return nil
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(info)
}
