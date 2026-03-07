package credential

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	credentialCreate "github.com/api7/a6/pkg/cmd/credential/create"
	credentialDelete "github.com/api7/a6/pkg/cmd/credential/delete"
	credentialGet "github.com/api7/a6/pkg/cmd/credential/get"
	credentialList "github.com/api7/a6/pkg/cmd/credential/list"
	credentialUpdate "github.com/api7/a6/pkg/cmd/credential/update"
)

func NewCmdCredential(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credential <command>",
		Short: "Manage APISIX consumer credentials",
	}

	cmd.AddCommand(credentialList.NewCmdList(f))
	cmd.AddCommand(credentialGet.NewCmdGet(f))
	cmd.AddCommand(credentialCreate.NewCmdCreate(f))
	cmd.AddCommand(credentialUpdate.NewCmdUpdate(f))
	cmd.AddCommand(credentialDelete.NewCmdDelete(f))

	return cmd
}
