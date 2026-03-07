package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/api7/a6/pkg/tableprinter"
)

type manager interface {
	List() ([]extension.Extension, error)
}

type Options struct {
	IO      *iostreams.IOStreams
	Manager manager
	Output  string
}

type extensionEntry struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
	Repo    string `json:"owner_repo" yaml:"owner_repo"`
}

func NewCmdList(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:      f.IOStreams,
		Manager: extension.NewManager(extension.DefaultExtensionsDir()),
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed extensions",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			output, _ := c.Flags().GetString("output")
			opts.Output = output
			return listRun(opts)
		},
	}

	return cmd
}

func listRun(opts *Options) error {
	exts, err := opts.Manager.List()
	if err != nil {
		return err
	}

	if len(exts) == 0 {
		fmt.Fprintln(opts.IO.Out, "No extensions installed.")
		return nil
	}

	if opts.Output == "json" || opts.Output == "yaml" {
		items := make([]extensionEntry, 0, len(exts))
		for _, ext := range exts {
			items = append(items, extensionEntry{
				Name:    ext.Name,
				Version: ext.Version,
				Repo:    ext.Owner + "/" + ext.Repo,
			})
		}
		return cmdutil.NewExporter(opts.Output, opts.IO.Out).Write(items)
	}

	tp := tableprinter.New(opts.IO.Out)
	tp.SetHeaders("NAME", "VERSION", "OWNER/REPO")
	for _, ext := range exts {
		tp.AddRow(ext.Name, ext.Version, ext.Owner+"/"+ext.Repo)
	}
	return tp.Render()
}
