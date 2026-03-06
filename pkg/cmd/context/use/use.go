package use

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

// Options holds the inputs for the use command.
type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)

	Name string
}

// NewCmdUse creates the `context use` command.
func NewCmdUse(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a different context",
		Long:  "Set the named context as the active connection context.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.Name = args[0]
			return useRun(opts)
		},
	}
}

func useRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	if err := cfg.SetCurrentContext(opts.Name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "✓ Switched to context %q.\n", opts.Name)
	return nil
}
