package list

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SSL certificates",
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

	body, err := client.Get("/apisix/admin/ssls", query)
	if err != nil {
		return fmt.Errorf("%s", cmdutil.FormatAPIError(err))
	}

	var resp api.ListResponse[api.SSL]
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	ssls := make([]api.SSL, len(resp.List))
	for i, item := range resp.List {
		ssls[i] = item.Value
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
		if len(ssls) == 0 {
			fmt.Fprintln(opts.IO.Out, "No SSL certificates found.")
			return nil
		}
		tp := tableprinter.New(opts.IO.Out)
		tp.SetHeaders("ID", "SNI", "STATUS", "TYPE", "VALIDITY", "CREATED")
		for _, s := range ssls {
			id := derefStr(s.ID)
			sni := formatSNI(s.SNI, s.SNIs)
			status := formatStatus(s.Status)
			typ := formatType(s.Type)
			validity := parseCertValidity(derefStr(s.Cert))
			created := formatCreated(s.CreateTime)
			tp.AddRow(id, sni, status, typ, validity, created)
		}
		return tp.Render()
	}

	return cmdutil.NewExporter(format, opts.IO.Out).Write(ssls)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatSNI(sni *string, snis []string) string {
	if len(snis) > 0 {
		joined := strings.Join(snis, ", ")
		if len(joined) > 50 {
			return joined[:47] + "..."
		}
		return joined
	}
	if sni != nil {
		return *sni
	}
	return ""
}

func formatStatus(status *int) string {
	if status == nil {
		return "enabled"
	}
	if *status == 1 {
		return "enabled"
	}
	return "disabled"
}

func formatType(typ *string) string {
	if typ == nil {
		return "server"
	}
	return *typ
}

// parseCertValidity extracts NotAfter from a PEM certificate.
func parseCertValidity(certPEM string) string {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "-"
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "-"
	}
	return cert.NotAfter.Format("2006-01-02")
}

func formatCreated(t *int64) string {
	if t == nil {
		return ""
	}
	return time.Unix(*t, 0).UTC().Format("2006-01-02 15:04:05")
}
