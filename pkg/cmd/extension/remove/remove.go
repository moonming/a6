package remove

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type manager interface {
	Remove(name string) error
}

type Options struct {
	IO      *iostreams.IOStreams
	Manager manager
	Force   bool
}

func NewCmdRemove(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:      f.IOStreams,
		Manager: extension.NewManager(extension.DefaultExtensionsDir()),
	}

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			force, _ := c.Flags().GetBool("force")
			opts.Force = force
			return removeRun(opts, args[0])
		},
	}

	return cmd
}

func removeRun(opts *Options, name string) error {
	if !opts.Force && opts.IO.IsStdinTTY() {
		fmt.Fprintf(opts.IO.ErrOut, "Remove extension %s? (y/N): ", name)
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

	if err := opts.Manager.Remove(name); err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.Out, "Removed extension %s\n", name)
	return nil
}
