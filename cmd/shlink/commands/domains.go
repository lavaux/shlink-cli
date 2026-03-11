package commands

import (
	"encoding/json"

	"github.com/lavaux/shlink-cli/internal/client"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/lavaux/shlink-cli/internal/output"
	"github.com/lavaux/shlink-cli/pkg/api"
	"github.com/spf13/cobra"
)

// NewDomainsCmd returns the "domains" command group.
func NewDomainsCmd(cfg func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Manage domains",
	}
	cmd.AddCommand(domainsListCmd(cfg))
	cmd.AddCommand(domainsSetRedirectsCmd(cfg))
	return cmd
}

func domainsListCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.Get("/domains", nil)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var list api.DomainList
			if err := json.Unmarshal(data, &list); err != nil {
				return err
			}

			headers := []string{"DOMAIN", "DEFAULT", "BASE REDIRECT", "404 REDIRECT", "INVALID URL REDIRECT"}
			rows := make([][]string, len(list.Domains.Data))
			for i, d := range list.Domains.Data {
				base, notFound, invalid := "", "", ""
				if d.Redirects != nil {
					if d.Redirects.BaseUrlRedirect != nil {
						base = *d.Redirects.BaseUrlRedirect
					}
					if d.Redirects.Regular404Redirect != nil {
						notFound = *d.Redirects.Regular404Redirect
					}
					if d.Redirects.InvalidShortURL != nil {
						invalid = *d.Redirects.InvalidShortURL
					}
				}
				rows[i] = []string{
					d.Domain,
					output.Bool(d.IsDefault),
					output.Truncate(base, 40),
					output.Truncate(notFound, 40),
					output.Truncate(invalid, 40),
				}
			}
			p.Table(headers, rows)
			return nil
		},
	}
}

func domainsSetRedirectsCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var (
		domain     string
		baseURL    string
		notFound   string
		invalidURL string
	)

	cmd := &cobra.Command{
		Use:   "set-redirects",
		Short: "Set redirect URLs for a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			body := map[string]interface{}{
				"domain": domain,
			}
			redirects := map[string]interface{}{}
			if cmd.Flags().Changed("base") {
				redirects["baseUrlRedirect"] = baseURL
			}
			if cmd.Flags().Changed("not-found") {
				redirects["regular404Redirect"] = notFound
			}
			if cmd.Flags().Changed("invalid") {
				redirects["invalidShortUrlRedirect"] = invalidURL
			}
			body["redirects"] = redirects

			data, err := c.Patch("/domains/redirects", body)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			p.Success("Redirects updated for domain %q", domain)
			return nil
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "Domain to configure (required)")
	cmd.Flags().StringVar(&baseURL, "base", "", "Redirect URL for the base domain path")
	cmd.Flags().StringVar(&notFound, "not-found", "", "Redirect URL for 404 responses")
	cmd.Flags().StringVar(&invalidURL, "invalid", "", "Redirect URL for invalid short codes")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}
