package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lavaux/shlink-cli/internal/client"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/lavaux/shlink-cli/internal/output"
	"github.com/lavaux/shlink-cli/pkg/api"

	"github.com/spf13/cobra"
)

// NewTagsCmd returns the "tags" command group.
func NewTagsCmd(cfg func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage tags",
	}
	cmd.AddCommand(tagsListCmd(cfg))
	cmd.AddCommand(tagsStatsCmd(cfg))
	cmd.AddCommand(tagsRenameCmd(cfg))
	cmd.AddCommand(tagsDeleteCmd(cfg))
	return cmd
}

func tagsListCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.Get("/tags", nil)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var list api.TagList
			if err := json.Unmarshal(data, &list); err != nil {
				return err
			}

			rows := make([][]string, len(list.Tags.Data))
			for i, t := range list.Tags.Data {
				rows[i] = []string{t}
			}
			p.Table([]string{"TAG"}, rows)
			return nil
		},
	}
}

func tagsStatsCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show tag statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.Get("/tags/stats", nil)
			if err != nil {
				return err
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var stats api.TagStatsList
			if err := json.Unmarshal(data, &stats); err != nil {
				return err
			}

			headers := []string{"TAG", "SHORT URLS", "VISITS (TOTAL)", "VISITS (BOTS)", "VISITS (NON-BOTS)"}
			rows := make([][]string, len(stats.Tags.Data))
			for i, t := range stats.Tags.Data {
				rows[i] = []string{
					t.Tag,
					strconv.Itoa(t.ShortURLsCount),
					strconv.Itoa(t.VisitsSummary.Total),
					strconv.Itoa(t.VisitsSummary.Bots),
					strconv.Itoa(t.VisitsSummary.NonBots),
				}
			}
			p.Table(headers, rows)
			return nil
		},
	}
}

func tagsRenameCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a tag",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			req := api.RenameTagRequest{OldName: args[0], NewName: args[1]}
			if _, err := c.Put("/tags", req); err != nil {
				return err
			}
			p.Success("Tag %q renamed to %q", args[0], args[1])
			return nil
		},
	}
}

func tagsDeleteCmd(cfgFn func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <tag>...",
		Aliases: []string{"rm"},
		Short:   "Delete one or more tags",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFn()
			if err != nil {
				return err
			}
			c := client.New(cfg)
			p := output.New(cfg.Output)

			for _, tag := range args {
				path := fmt.Sprintf("/tags?tags[]=%s", tag)
				if err := c.Delete(path); err != nil {
					return fmt.Errorf("deleting tag %q: %w", tag, err)
				}
				p.Success("Tag %q deleted", tag)
			}
			return nil
		},
	}
}
