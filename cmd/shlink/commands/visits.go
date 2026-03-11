package commands

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/lavaux/shlink-cli/internal/client"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/lavaux/shlink-cli/internal/output"
	"github.com/lavaux/shlink-cli/pkg/api"

	"github.com/spf13/cobra"
)

// NewVisitsCmd returns the "visits" command group.
func NewVisitsCmd(cfg func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visits",
		Short: "View visit statistics",
	}
	cmd.AddCommand(visitsGlobalCmd(cfg))
	cmd.AddCommand(visitsListCmd(cfg))
	cmd.AddCommand(visitsOrphanCmd(cfg))
	cmd.AddCommand(visitsByTagCmd(cfg))
	cmd.AddCommand(visitsByDomainCmd(cfg))
	return cmd
}

// shared visit query flags
type visitQueryFlags struct {
	StartDate   string
	EndDate     string
	ExcludeBots bool
	Page        int
	PerPage     int
}

func (f *visitQueryFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.StartDate, "start-date", "", "Filter visits from this date (RFC 3339)")
	cmd.Flags().StringVar(&f.EndDate, "end-date", "", "Filter visits to this date (RFC 3339)")
	cmd.Flags().BoolVar(&f.ExcludeBots, "exclude-bots", false, "Exclude bot visits")
	cmd.Flags().IntVar(&f.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&f.PerPage, "per-page", 20, "Items per page")
}

func (f *visitQueryFlags) values() url.Values {
	q := url.Values{}
	if f.StartDate != "" {
		q.Set("startDate", f.StartDate)
	}
	if f.EndDate != "" {
		q.Set("endDate", f.EndDate)
	}
	if f.ExcludeBots {
		q.Set("excludeBots", "true")
	}
	q.Set("page", fmt.Sprintf("%d", f.Page))
	q.Set("itemsPerPage", fmt.Sprintf("%d", f.PerPage))
	return q
}

func printVisitList(p *output.Printer, cfg *config.Config, data []byte) error {
	if cfg.Output == "json" {
		return p.JSON(data)
	}

	var list api.VisitList
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	headers := []string{"DATE", "TYPE", "REFERER", "COUNTRY", "CITY", "BOT", "VISITED URL"}
	rows := make([][]string, len(list.Visits.Data))
	for i, v := range list.Visits.Data {
		country, city := "", ""
		if v.VisitLocation != nil {
			country = v.VisitLocation.CountryName
			city = v.VisitLocation.CityName
		}
		rows[i] = []string{
			v.Date[:10],
			v.Type,
			output.Truncate(v.Referer, 30),
			country,
			city,
			output.Bool(v.PotentialBot),
			output.Truncate(v.VisitedURL, 40),
		}
	}
	p.Table(headers, rows)
	pg := list.Visits.Pagination
	p.Line("\nPage %d of %d  (%d total visits)", pg.CurrentPage, pg.PagesCount, pg.TotalItems)
	return nil
}

func visitsGlobalCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "global",
		Short: "Show global visit counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.Get("/visits", nil)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var g api.GlobalVisits
			if err := json.Unmarshal(data, &g); err != nil {
				return err
			}

			rows := [][]string{
				{"Non-orphan visits (total)", fmt.Sprintf("%d", g.Visits.NonOrphanVisits.Total)},
				{"Non-orphan visits (bots)", fmt.Sprintf("%d", g.Visits.NonOrphanVisits.Bots)},
				{"Non-orphan visits (non-bots)", fmt.Sprintf("%d", g.Visits.NonOrphanVisits.NonBots)},
				{"Orphan visits (total)", fmt.Sprintf("%d", g.Visits.OrphanVisits.Total)},
				{"Orphan visits (bots)", fmt.Sprintf("%d", g.Visits.OrphanVisits.Bots)},
				{"Orphan visits (non-bots)", fmt.Sprintf("%d", g.Visits.OrphanVisits.NonBots)},
			}
			p.Table([]string{"METRIC", "COUNT"}, rows)
			return nil
		},
	}
}

func visitsListCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var qf visitQueryFlags
	var domain string

	cmd := &cobra.Command{
		Use:   "list <shortCode>",
		Short: "List visits for a short URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			q := qf.values()
			if domain != "" {
				q.Set("domain", domain)
			}

			path := fmt.Sprintf("/short-urls/%s/visits", args[0])
			data, err := c.Get(path, q)
			if err != nil {
				return err
			}
			return printVisitList(p, cfg, data)
		},
	}
	qf.register(cmd)
	cmd.Flags().StringVar(&domain, "domain", "", "Domain qualifier")
	return cmd
}

func visitsOrphanCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var qf visitQueryFlags

	cmd := &cobra.Command{
		Use:   "orphan",
		Short: "List orphan visits (no matching short URL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.Get("/visits/orphan", qf.values())
			if err != nil {
				return err
			}
			return printVisitList(p, cfg, data)
		},
	}
	qf.register(cmd)
	return cmd
}

func visitsByTagCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var qf visitQueryFlags

	cmd := &cobra.Command{
		Use:   "tag <tagName>",
		Short: "List visits for a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			path := fmt.Sprintf("/tags/%s/visits", url.PathEscape(args[0]))
			data, err := c.Get(path, qf.values())
			if err != nil {
				return err
			}
			return printVisitList(p, cfg, data)
		},
	}
	qf.register(cmd)
	return cmd
}

func visitsByDomainCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var qf visitQueryFlags

	cmd := &cobra.Command{
		Use:   "domain <domainName>",
		Short: "List visits for a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			path := fmt.Sprintf("/domains/%s/visits", url.PathEscape(args[0]))
			data, err := c.Get(path, qf.values())
			if err != nil {
				return err
			}
			return printVisitList(p, cfg, data)
		},
	}
	qf.register(cmd)
	return cmd
}
