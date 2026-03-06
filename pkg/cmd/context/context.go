package context

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/context/create"
	"github.com/api7/a6/pkg/cmd/context/current"
	deletecmd "github.com/api7/a6/pkg/cmd/context/delete"
	"github.com/api7/a6/pkg/cmd/context/list"
	"github.com/api7/a6/pkg/cmd/context/use"
)

// NewCmdContext creates the `context` parent command.
func NewCmdContext(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context <command>",
		Short: "Manage APISIX connection contexts",
		Long:  "Commands for creating, switching, listing, and deleting APISIX connection contexts.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(use.NewCmdUse(f))
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))
	cmd.AddCommand(current.NewCmdCurrent(f))

	return cmd
}
