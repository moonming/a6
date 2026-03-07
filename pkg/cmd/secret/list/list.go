package list

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		Use:   "list",
		Short: "List secret managers",
		Args:  cobra.NoArgs,
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

	httpClient, err := opts.Client()
	if err != nil {
		return err
	}

	client := api.NewClient(httpClient, cfg.BaseURL())

	query := map[string]string{
		"page":      fmt.Sprintf("%d", opts.Page),
		"page_size": fmt.Sprintf("%d", opts.PageSize),
	}

	body, err := client.Get("/apisix/admin/secrets", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.Secret]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	secrets := make([]api.Secret, len(resp.List))
	for i, item := range resp.List {
		secrets[i] = item.Value
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
		if len(resp.List) == 0 {
			fmt.Fprintln(opts.IO.Out, "No secrets found.")
			return nil
		}

		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "MANAGER", "URI/REGION", "CREATED")
		for _, item := range resp.List {
			id, manager := parseSecretKey(item.Key)
			endpoint := secretEndpoint(manager, item.Value)
			created := ""
			if item.Value.CreateTime != nil {
				created = time.Unix(*item.Value.CreateTime, 0).Format("2006-01-02 15:04:05")
			}
			tp.AddRow(id, manager, endpoint, created)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(secrets)
}

func parseSecretKey(key string) (string, string) {
	parts := strings.Split(strings.Trim(key, "/"), "/")
	if len(parts) < 2 {
		return key, ""
	}
	id := strings.Join(parts[len(parts)-2:], "/")
	return id, parts[len(parts)-2]
}

func secretEndpoint(manager string, s api.Secret) string {
	if manager == "vault" {
		return derefStr(s.URI)
	}
	if manager == "aws" {
		return derefStr(s.Region)
	}
	if manager == "gcp" {
		return "gcp"
	}
	if s.URI != nil {
		return *s.URI
	}
	if s.Region != nil {
		return *s.Region
	}
	return ""
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
