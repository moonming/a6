package config

import (
	"github.com/spf13/cobra"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmd/config/diff"
	"github.com/api7/a6/pkg/cmd/config/dump"
	"github.com/api7/a6/pkg/cmd/config/sync"
	"github.com/api7/a6/pkg/cmd/config/validate"
)

func NewCmdConfig(f *cmd.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage declarative APISIX configuration",
	}

	cmd.AddCommand(dump.NewCmdDump(f))
	cmd.AddCommand(diff.NewCmdDiff(f))
	cmd.AddCommand(sync.NewCmdSync(f))
	cmd.AddCommand(validate.NewCmdValidate(f))

	return cmd
}
