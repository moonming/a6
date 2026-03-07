package upgrade

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type manager interface {
	Find(name string) (*extension.Extension, error)
	Upgrade(name string) (*extension.Extension, error)
	UpgradeAll() ([]extension.Extension, error)
}

type Options struct {
	IO      *iostreams.IOStreams
	Manager manager
	All     bool
}

func NewCmdUpgrade(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:      f.IOStreams,
		Manager: extension.NewManager(extension.DefaultExtensionsDir()),
	}

	cmd := &cobra.Command{
		Use:   "upgrade [name]",
		Short: "Upgrade installed extensions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.All, _ = c.Flags().GetBool("all")
			hasName := len(args) == 1
			if opts.All && hasName {
				return fmt.Errorf("name argument and --all are mutually exclusive")
			}
			if !opts.All && !hasName {
				return fmt.Errorf("must specify an extension name or --all")
			}
			if opts.All {
				return upgradeAllRun(opts)
			}
			return upgradeRun(opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Upgrade all installed extensions")

	return cmd
}

func upgradeRun(opts *Options, name string) error {
	current, err := opts.Manager.Find(name)
	if err != nil {
		return err
	}
	updated, err := opts.Manager.Upgrade(name)
	if err != nil {
		return err
	}
	if current.Version == updated.Version {
		fmt.Fprintf(opts.IO.Out, "Extension %s is already up to date\n", name)
		return nil
	}
	fmt.Fprintf(opts.IO.Out, "Upgraded %s from v%s to v%s\n", name, current.Version, updated.Version)
	return nil
}

func upgradeAllRun(opts *Options) error {
	upgraded, err := opts.Manager.UpgradeAll()
	if err != nil {
		return err
	}
	if len(upgraded) == 0 {
		fmt.Fprintln(opts.IO.Out, "All extensions are already up to date")
		return nil
	}
	for _, ext := range upgraded {
		fmt.Fprintf(opts.IO.Out, "Upgraded %s to v%s\n", ext.Name, ext.Version)
	}
	return nil
}
