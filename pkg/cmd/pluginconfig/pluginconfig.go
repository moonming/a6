package pluginconfig

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	pluginconfigCreate "github.com/api7/a6/pkg/cmd/pluginconfig/create"
	pluginconfigDelete "github.com/api7/a6/pkg/cmd/pluginconfig/delete"
	pluginconfigExport "github.com/api7/a6/pkg/cmd/pluginconfig/export"
	pluginconfigGet "github.com/api7/a6/pkg/cmd/pluginconfig/get"
	pluginconfigList "github.com/api7/a6/pkg/cmd/pluginconfig/list"
	pluginconfigUpdate "github.com/api7/a6/pkg/cmd/pluginconfig/update"
)

// NewCmdPluginConfig creates parent command.
func NewCmdPluginConfig(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin-config <command>",
		Short: "Manage APISIX plugin configs",
	}

	cmd.AddCommand(pluginconfigList.NewCmdList(f))
	cmd.AddCommand(pluginconfigGet.NewCmdGet(f))
	cmd.AddCommand(pluginconfigCreate.NewCmdCreate(f))
	cmd.AddCommand(pluginconfigUpdate.NewCmdUpdate(f))
	cmd.AddCommand(pluginconfigDelete.NewCmdDelete(f))
	cmd.AddCommand(pluginconfigExport.NewCmdExport(f))

	return cmd
}
