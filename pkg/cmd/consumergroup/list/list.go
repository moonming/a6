package list

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/api"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/cmdutil"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/api7/a6/pkg/tableprinter"
)

type Options struct {
	IO     *iostreams.IOStreams
	Client func() (*http.Client, error)
	Config func() (config.Config, error)

	Page     int
	PageSize int
	Output   string
}

func NewCmdList(f *cmd.Factory) *cobra.Command {
	opts := &Options{
		IO:     f.IOStreams,
		Client: f.HttpClient,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List consumer groups",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Page size")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output format: json, yaml, table")

	return cmd
}

func listRun(opts *Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	baseURL := cfg.BaseURL()

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, baseURL)

	query := map[string]string{
		"page":      fmt.Sprintf("%d", opts.Page),
		"page_size": fmt.Sprintf("%d", opts.PageSize),
	}

	body, err := client.Get("/apisix/admin/consumer_groups", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.ConsumerGroup]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	consumerGroups := make([]api.ConsumerGroup, len(resp.List))
	for i, item := range resp.List {
		consumerGroups[i] = item.Value
	}

	format := opts.Output
	if format == "" {
		if opts.IO.IsStdoutTTY() {
			format = "table"
		} else {
			format = "json"
		}
	}

	if format == "table" {
		if len(consumerGroups) == 0 {
			fmt.Fprintln(opts.IO.Out, "No consumer groups found.")
			return nil
		}
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "NAME", "PLUGINS", "CREATED")
		for _, cg := range consumerGroups {
			id := derefStr(cg.ID)
			name := derefStr(cg.Name)
			plugins := formatPlugins(cg.Plugins)
			created := ""
			if cg.CreateTime != nil {
				created = time.Unix(*cg.CreateTime, 0).Format("2006-01-02 15:04:05")
			}
			tp.AddRow(id, name, plugins, created)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(consumerGroups)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatPlugins(plugins map[string]interface{}) string {
	count := len(plugins)
	if count == 0 {
		return ""
	}
	names := make([]string, 0, count)
	for name := range plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	if count <= 3 {
		return strings.Join(names, ",")
	}
	return fmt.Sprintf("%d plugins (%s...)", count, strings.Join(names[:3], ","))
}
