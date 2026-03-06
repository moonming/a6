package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/api7/a6/pkg/tableprinter"
)

// Options holds the inputs for the list command.
type Options struct {
	IO     *iostreams.IOStreams
	Config func() (config.Config, error)
}

// NewCmdList creates the `context list` command.
func NewCmdList(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all contexts",
		Long:    "List all configured connection contexts.",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			output, _ := c.Flags().GetString("output")
			return listRun(opts, output)
		},
	}
}

// contextEntry is the JSON/YAML output structure for a context.
type contextEntry struct {
	Name    string `json:"name" yaml:"name"`
	Server  string `json:"server" yaml:"server"`
	Current bool   `json:"current" yaml:"current"`
}

func listRun(opts *Options, output string) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	contexts := cfg.Contexts()
	current := cfg.CurrentContext()

	if len(contexts) == 0 {
		fmt.Fprintln(opts.IO.Out, "No contexts configured. Use \"a6 context create\" to add one.")
		return nil
	}

	// Structured output (JSON/YAML).
	if output == "json" || output == "yaml" {
		entries := make([]contextEntry, len(contexts))
		for i, ctx := range contexts {
			entries[i] = contextEntry{
				Name:    ctx.Name,
				Server:  ctx.Server,
				Current: ctx.Name == current,
			}
		}
		return cmdutil.NewExporter(output, opts.IO.Out).Write(entries)
	}

	// Table output.
	tp := tableprinter.New(opts.IO.Out)
	tp.SetHeaders("NAME", "SERVER", "CURRENT")
	for _, ctx := range contexts {
		marker := ""
		if ctx.Name == current {
			marker = "*"
		}
		tp.AddRow(ctx.Name, ctx.Server, marker)
	}
	return tp.Render()
}
