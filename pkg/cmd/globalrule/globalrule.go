package globalrule

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	globalruleCreate "github.com/api7/a6/pkg/cmd/globalrule/create"
	globalruleDelete "github.com/api7/a6/pkg/cmd/globalrule/delete"
	globalruleGet "github.com/api7/a6/pkg/cmd/globalrule/get"
	globalruleList "github.com/api7/a6/pkg/cmd/globalrule/list"
	globalruleUpdate "github.com/api7/a6/pkg/cmd/globalrule/update"
)

func NewCmdGlobalRule(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-rule <command>",
		Short: "Manage APISIX global rules",
	}

	cmd.AddCommand(globalruleList.NewCmdList(f))
	cmd.AddCommand(globalruleGet.NewCmdGet(f))
	cmd.AddCommand(globalruleCreate.NewCmdCreate(f))
	cmd.AddCommand(globalruleUpdate.NewCmdUpdate(f))
	cmd.AddCommand(globalruleDelete.NewCmdDelete(f))

	return cmd
}
