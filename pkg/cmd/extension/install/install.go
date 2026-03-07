package install

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type manager interface {
	Install(ownerRepo string) (*extension.Extension, error)
}

type Options struct {
	IO      *iostreams.IOStreams
	Manager manager
}

func NewCmdInstall(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:      f.IOStreams,
		Manager: extension.NewManager(extension.DefaultExtensionsDir()),
	}

	cmd := &cobra.Command{
		Use:   "install <owner/repo>",
		Short: "Install an extension from GitHub Releases",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if strings.Count(args[0], "/") != 1 {
				return fmt.Errorf("owner/repo must be in the format <owner>/<repo>")
			}
			return installRun(opts, args[0])
		},
	}

	return cmd
}

func installRun(opts *Options, ownerRepo string) error {
	ext, err := opts.Manager.Install(ownerRepo)
	if err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.Out, "Installed extension %s v%s\n", ext.Name, ext.Version)
	return nil
}
