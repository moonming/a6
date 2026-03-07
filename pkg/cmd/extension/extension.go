package extension

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/extension/install"
	"github.com/api7/a6/pkg/cmd/extension/list"
	"github.com/api7/a6/pkg/cmd/extension/remove"
	"github.com/api7/a6/pkg/cmd/extension/upgrade"
)

func NewCmdExtension(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension <command>",
		Short:   "Manage a6 CLI extensions",
		Long:    "Install, manage, and remove CLI extensions that add new commands to a6.",
		Aliases: []string{"ext"},
	}

	cmd.AddCommand(install.NewCmdInstall(f))
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(upgrade.NewCmdUpgrade(f))
	cmd.AddCommand(remove.NewCmdRemove(f))

	return cmd
}
