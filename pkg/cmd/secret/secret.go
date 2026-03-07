package secret

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	secretCreate "github.com/api7/a6/pkg/cmd/secret/create"
	secretDelete "github.com/api7/a6/pkg/cmd/secret/delete"
	secretGet "github.com/api7/a6/pkg/cmd/secret/get"
	secretList "github.com/api7/a6/pkg/cmd/secret/list"
	secretUpdate "github.com/api7/a6/pkg/cmd/secret/update"
)

func NewCmdSecret(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret <command>",
		Short: "Manage APISIX secret managers",
	}

	cmd.AddCommand(secretList.NewCmdList(f))
	cmd.AddCommand(secretGet.NewCmdGet(f))
	cmd.AddCommand(secretCreate.NewCmdCreate(f))
	cmd.AddCommand(secretUpdate.NewCmdUpdate(f))
	cmd.AddCommand(secretDelete.NewCmdDelete(f))

	return cmd
}
