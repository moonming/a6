package ssl

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/ssl/create"
	deletecmd "github.com/api7/a6/pkg/cmd/ssl/delete"
	"github.com/api7/a6/pkg/cmd/ssl/export"
	"github.com/api7/a6/pkg/cmd/ssl/get"
	"github.com/api7/a6/pkg/cmd/ssl/list"
	"github.com/api7/a6/pkg/cmd/ssl/update"
)

// NewCmdSSL creates the `ssl` parent command.
func NewCmdSSL(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl <command>",
		Short: "Manage APISIX SSL certificates",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))
	cmd.AddCommand(export.NewCmdExport(f))

	return cmd
}
