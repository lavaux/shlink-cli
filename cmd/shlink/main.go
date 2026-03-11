// Package main is the entry point for the shlink CLI.
package main

import (
	"fmt"
	"os"

	"github.com/lavaux/shlink-cli/cmd/shlink/commands"
	"github.com/lavaux/shlink-cli/internal/config"
	"github.com/spf13/cobra"
)

// globalFlags holds persistent flag values set on the root command.
type globalFlags struct {
	Server  string
	APIKey  string
	Output  string
	Version string
}

var flags globalFlags

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "shlink",
	Short: "A CLI client for the Shlink self-hosted URL shortener",
	Long: `shlink-cli lets you manage your Shlink instance from the command line.

Configuration (in order of priority):
  1. Command-line flags   --server, --api-key
  2. Environment variables SHLINK_SERVER, SHLINK_API_KEY

Examples:
  shlink urls list
  shlink urls create https://example.com --slug my-link
  shlink urls get my-link
  shlink tags list
  shlink visits list my-link
  shlink domains list
  shlink health`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.Server, "server", "", "Shlink server URL (overrides SHLINK_SERVER)")
	rootCmd.PersistentFlags().StringVar(&flags.APIKey, "api-key", "", "Shlink API key (overrides SHLINK_API_KEY)")
	rootCmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().StringVar(&flags.Version, "api-version", "3", "Shlink REST API version")

	// Build config factory that subcommands will call.
	cfgFactory := func() (*config.Config, error) {
		cfg := config.FromEnv()
		cfg.Merge(flags.Server, flags.APIKey, flags.Output, flags.Version)
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	rootCmd.AddCommand(commands.NewURLsCmd(cfgFactory))
	rootCmd.AddCommand(commands.NewTagsCmd(cfgFactory))
	rootCmd.AddCommand(commands.NewVisitsCmd(cfgFactory))
	rootCmd.AddCommand(commands.NewDomainsCmd(cfgFactory))
	rootCmd.AddCommand(commands.NewHealthCmd(cfgFactory))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
