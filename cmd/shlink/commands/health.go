package commands

import (
	"encoding/json"
	"fmt"

	"github.com/lavaux/shlink-cli/internal/client"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/lavaux/shlink-cli/internal/output"
	"github.com/lavaux/shlink-cli/pkg/api"

	"github.com/spf13/cobra"
)

// NewHealthCmd returns the "health" command.
func NewHealthCmd(cfg func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check the health of the Shlink server",
		// Health does not require an API key, but we still need the server URL.
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfg()
			if err != nil {
				// Allow missing API key for health check.
				if cfg == nil {
					return fmt.Errorf("server URL is required (set SHLINK_SERVER or --server)")
				}
				// If only API key is missing, continue.
			}

			c := client.New(cfg)
			p := output.New(cfg.Output)

			data, err := c.GetHealth()
			if err != nil {
				return fmt.Errorf("health check failed: %w", err)
			}

			if cfg.Output == "json" {
				return p.JSON(data)
			}

			var h api.Health
			if err := json.Unmarshal(data, &h); err != nil {
				return err
			}

			rows := [][]string{
				{"Status", h.Status},
				{"Version", h.Version},
			}
			p.Table([]string{"FIELD", "VALUE"}, rows)
			return nil
		},
	}
}
