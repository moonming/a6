package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

// Options holds the inputs for the create command.
type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)

	Name   string
	Server string
	APIKey string
}

// NewCmdCreate creates the `context create` command.
func NewCmdCreate(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new connection context",
		Long:  "Create a new named context for connecting to an APISIX instance.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.Name = args[0]
			return createRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Server, "server", "", "APISIX Admin API server URL (required)")
	cmd.Flags().StringVar(&opts.APIKey, "api-key", "", "Admin API key")
	_ = cmd.MarkFlagRequired("server")

	return cmd
}

func createRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	ctx := config.Context{
		Name:   opts.Name,
		Server: opts.Server,
		APIKey: opts.APIKey,
	}

	hadContexts := len(cfg.Contexts()) > 0

	if err := cfg.AddContext(ctx); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "✓ Context %q created and saved.\n", opts.Name)

	// If this is the first context, it was auto-set as current.
	if !hadContexts {
		fmt.Fprintf(opts.IO.Out, "✓ Context %q set as current context.\n", opts.Name)
	}

	return nil
}
