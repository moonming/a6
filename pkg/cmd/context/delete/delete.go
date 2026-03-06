package delete

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

// Options holds the inputs for the delete command.
type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)

	Name string
}

// NewCmdDelete creates the `context delete` command.
func NewCmdDelete(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a context",
		Long:    "Delete a named connection context from the configuration.",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.Name = args[0]
			force, _ := c.Flags().GetBool("force")
			return deleteRun(opts, force)
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt when deleting the active context")

	return cmd
}

func deleteRun(opts *Options, force bool) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	// If deleting the active context, prompt for confirmation.
	if cfg.CurrentContext() == opts.Name && !force {
		if opts.IO.IsStdinTTY() {
			fmt.Fprintf(opts.IO.ErrOut, "Context %q is the current active context. Delete it? (y/N): ", opts.Name)
			reader := bufio.NewReader(opts.IO.In)
			answer, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Fprintln(opts.IO.ErrOut, "Aborted.")
				return nil
			}
		}
	}

	if err := cfg.RemoveContext(opts.Name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "✓ Context %q deleted.\n", opts.Name)
	return nil
}
