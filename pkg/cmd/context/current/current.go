package current

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

// Options holds the inputs for the current command.
type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)
}

// NewCmdCurrent creates the `context current` command.
func NewCmdCurrent(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	return &cobra.Command{
		Use:   "current",
		Short: "Show the current active context",
		Long:  "Display the name of the currently active connection context.",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return currentRun(opts)
		},
	}
}

func currentRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	name := cfg.CurrentContext()
	if name == "" {
		return fmt.Errorf("no current context configured; use \"a6 context create\" to add one")
	}

	fmt.Fprintln(opts.IO.Out, name)
	return nil
}
