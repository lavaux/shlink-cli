package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lavaux/shlink-cli/internal/client"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/lavaux/shlink-cli/internal/output"
	"github.com/lavaux/shlink-cli/pkg/api"

	"github.com/spf13/cobra"
)

// NewURLsCmd returns the "urls" command group.
func NewURLsCmd(cfg func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "urls",
		Aliases: []string{"url", "short-urls"},
		Short:   "Manage short URLs",
	}
	cmd.AddCommand(urlsListCmd(cfg))
	cmd.AddCommand(urlsGetCmd(cfg))
	cmd.AddCommand(urlsCreateCmd(cfg))
	cmd.AddCommand(urlsEditCmd(cfg))
	cmd.AddCommand(urlsDeleteCmd(cfg))
	return cmd
}

// ---- list ----

func urlsListCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var page int
	var pageSize int
	var searchTerm string
	var tags []string
	var orderBy string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all short URLs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			q := url.Values{}
			q.Set("page", strconv.Itoa(page))
			q.Set("itemsPerPage", strconv.Itoa(pageSize))
			if searchTerm != "" {
				q.Set("searchTerm", searchTerm)
			}
			for _, t := range tags {
				q.Add("tags[]", t)
			}
			if orderBy != "" {
				q.Set("orderBy", orderBy)
			}

			data, err := c.Get("/short-urls", q)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var list api.ShortURLList
			if err := json.Unmarshal(data, &list); err != nil {
				return err
			}

			headers := []string{"SHORT CODE", "SHORT URL", "LONG URL", "VISITS", "TAGS", "CREATED"}
			rows := make([][]string, len(list.ShortURLs.Data))
			for i, u := range list.ShortURLs.Data {
				rows[i] = []string{
					u.ShortCode,
					u.ShortURL,
					output.Truncate(u.LongURL, 60),
					strconv.Itoa(u.VisitsSummary.Total),
					strings.Join(u.Tags, ", "),
					u.DateCreated[:10],
				}
			}
			p.Table(headers, rows)
			pg := list.ShortURLs.Pagination
			p.Line("\nPage %d of %d  (%d total)", pg.CurrentPage, pg.PagesCount, pg.TotalItems)
			return nil
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&pageSize, "per-page", 10, "Items per page")
	cmd.Flags().StringVar(&searchTerm, "search", "", "Search term")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tag (repeatable)")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order field, e.g. dateCreated-DESC")
	return cmd
}

// ---- get ----

func urlsGetCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "get <shortCode>",
		Short: "Get details of a short URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			path := fmt.Sprintf("/short-urls/%s", args[0])
			q := url.Values{}
			if domain != "" {
				q.Set("domain", domain)
			}

			data, err := c.Get(path, q)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var u api.ShortURL
			if err := json.Unmarshal(data, &u); err != nil {
				return err
			}

			title := ""
			if u.Title != nil {
				title = *u.Title
			}
			d := ""
			if u.Domain != nil {
				d = *u.Domain
			}

			rows := [][]string{
				{"Short Code", u.ShortCode},
				{"Short URL", u.ShortURL},
				{"Long URL", u.LongURL},
				{"Title", title},
				{"Domain", d},
				{"Tags", strings.Join(u.Tags, ", ")},
				{"Created", u.DateCreated},
				{"Crawlable", output.Bool(u.Crawlable)},
				{"Forward Query", output.Bool(u.ForwardQuery)},
				{"Visits (total)", strconv.Itoa(u.VisitsSummary.Total)},
				{"Visits (bots)", strconv.Itoa(u.VisitsSummary.Bots)},
				{"Visits (non-bots)", strconv.Itoa(u.VisitsSummary.NonBots)},
			}
			p.Table([]string{"FIELD", "VALUE"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "Domain (if short code is not unique)")
	return cmd
}

// ---- create ----

func urlsCreateCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var (
		slug         string
		title        string
		tags         []string
		domain       string
		maxVisits    int
		validSince   string
		validUntil   string
		crawlable    bool
		forwardQuery bool
		findIfExists bool
	)

	cmd := &cobra.Command{
		Use:   "create <longURL>",
		Short: "Create a new short URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			req := api.CreateShortURLRequest{
				LongURL:      args[0],
				ShortCode:    slug,
				Title:        title,
				Tags:         tags,
				Domain:       domain,
				ValidSince:   validSince,
				ValidUntil:   validUntil,
				FindIfExists: findIfExists,
			}
			if cmd.Flags().Changed("max-visits") {
				req.MaxVisits = &maxVisits
			}
			if cmd.Flags().Changed("crawlable") {
				req.Crawlable = &crawlable
			}
			if cmd.Flags().Changed("forward-query") {
				req.ForwardQuery = &forwardQuery
			}

			data, err := c.Post("/short-urls", req)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var u api.ShortURL
			if err := json.Unmarshal(data, &u); err != nil {
				return err
			}
			p.Success("Short URL created: %s → %s", u.ShortURL, u.LongURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Custom slug / short code")
	cmd.Flags().StringVar(&title, "title", "", "Title for the URL")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags (repeatable)")
	cmd.Flags().StringVar(&domain, "domain", "", "Custom domain")
	cmd.Flags().IntVar(&maxVisits, "max-visits", 0, "Maximum number of visits")
	cmd.Flags().StringVar(&validSince, "valid-since", "", "Valid since date (RFC 3339)")
	cmd.Flags().StringVar(&validUntil, "valid-until", "", "Valid until date (RFC 3339)")
	cmd.Flags().BoolVar(&crawlable, "crawlable", false, "Allow crawling by bots")
	cmd.Flags().BoolVar(&forwardQuery, "forward-query", false, "Forward query params to long URL")
	cmd.Flags().BoolVar(&findIfExists, "find-if-exists", false, "Return existing URL if long URL already shortened")
	return cmd
}

// ---- edit ----

func urlsEditCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var (
		longURL         string
		title           string
		tags            []string
		domain          string
		maxVisits       int
		clearMaxVisits  bool
		validSince      string
		clearValidSince bool
		validUntil      string
		clearValidUntil bool
		crawlable       bool
		forwardQuery    bool
	)

	cmd := &cobra.Command{
		Use:   "edit <shortCode>",
		Short: "Edit an existing short URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			req := api.EditShortURLRequest{}
			if longURL != "" {
				req.LongURL = longURL
			}
			if cmd.Flags().Changed("title") {
				req.Title = &title
			}
			if cmd.Flags().Changed("tag") {
				req.Tags = tags
			}
			if cmd.Flags().Changed("max-visits") {
				req.MaxVisits = &maxVisits
			}
			if clearMaxVisits {
				zero := 0
				req.MaxVisits = &zero
			}
			if cmd.Flags().Changed("valid-since") {
				req.ValidSince = &validSince
			}
			if clearValidSince {
				empty := ""
				req.ValidSince = &empty
			}
			if cmd.Flags().Changed("valid-until") {
				req.ValidUntil = &validUntil
			}
			if clearValidUntil {
				empty := ""
				req.ValidUntil = &empty
			}
			if cmd.Flags().Changed("crawlable") {
				req.Crawlable = &crawlable
			}
			if cmd.Flags().Changed("forward-query") {
				req.ForwardQuery = &forwardQuery
			}

			path := fmt.Sprintf("/short-urls/%s", args[0])
			if domain != "" {
				path += "?domain=" + url.QueryEscape(domain)
			}

			data, err := c.Patch(path, req)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var u api.ShortURL
			if err := json.Unmarshal(data, &u); err != nil {
				return err
			}
			p.Success("Short URL %s updated", u.ShortCode)
			return nil
		},
	}

	cmd.Flags().StringVar(&longURL, "long-url", "", "New destination URL")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Replace tags (repeatable)")
	cmd.Flags().StringVar(&domain, "domain", "", "Domain qualifier")
	cmd.Flags().IntVar(&maxVisits, "max-visits", 0, "New max visits limit")
	cmd.Flags().BoolVar(&clearMaxVisits, "clear-max-visits", false, "Remove max visits limit")
	cmd.Flags().StringVar(&validSince, "valid-since", "", "New valid-since date")
	cmd.Flags().BoolVar(&clearValidSince, "clear-valid-since", false, "Remove valid-since date")
	cmd.Flags().StringVar(&validUntil, "valid-until", "", "New valid-until date")
	cmd.Flags().BoolVar(&clearValidUntil, "clear-valid-until", false, "Remove valid-until date")
	cmd.Flags().BoolVar(&crawlable, "crawlable", false, "Set crawlable flag")
	cmd.Flags().BoolVar(&forwardQuery, "forward-query", false, "Set forward-query flag")
	return cmd
}

// ---- delete ----

func urlsDeleteCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:     "delete <shortCode>",
		Aliases: []string{"rm"},
		Short:   "Delete a short URL",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			path := fmt.Sprintf("/short-urls/%s", args[0])
			if domain != "" {
				path += "?domain=" + url.QueryEscape(domain)
			}

			if err := c.Delete(path); err != nil {
				return err
			}
			p.Success("Short URL %q deleted", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "Domain qualifier")
	return cmd
}
