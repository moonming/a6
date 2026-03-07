package consumergroup

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	consumergroupCreate "github.com/api7/a6/pkg/cmd/consumergroup/create"
	consumergroupDelete "github.com/api7/a6/pkg/cmd/consumergroup/delete"
	consumergroupGet "github.com/api7/a6/pkg/cmd/consumergroup/get"
	consumergroupList "github.com/api7/a6/pkg/cmd/consumergroup/list"
	consumergroupUpdate "github.com/api7/a6/pkg/cmd/consumergroup/update"
)

func NewCmdConsumerGroup(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer-group <command>",
		Short: "Manage APISIX consumer groups",
	}

	cmd.AddCommand(consumergroupList.NewCmdList(f))
	cmd.AddCommand(consumergroupGet.NewCmdGet(f))
	cmd.AddCommand(consumergroupCreate.NewCmdCreate(f))
	cmd.AddCommand(consumergroupUpdate.NewCmdUpdate(f))
	cmd.AddCommand(consumergroupDelete.NewCmdDelete(f))

	return cmd
}
