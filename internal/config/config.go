// Package config manages CLI configuration from environment variables and flags.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the runtime configuration for the CLI.
type Config struct {
	ServerURL string
	APIKey    string
	Output    string // "table", "json", "plain"
	Version   string // API version, default "3"
}

// FromEnv loads configuration from environment variables.
// Environment variables:
//   - SHLINK_SERVER  — base URL of the Shlink instance (e.g. https://s.example.com)
//   - SHLINK_API_KEY — API key for authentication
func FromEnv() *Config {
	return &Config{
		ServerURL: os.Getenv("SHLINK_SERVER"),
		APIKey:    os.Getenv("SHLINK_API_KEY"),
		Output:    "table",
		Version:   "3",
	}
}

// Merge applies overrides from command-line flags (non-empty values win).
func (c *Config) Merge(server, apiKey, output, version string) {
	if server != "" {
		c.ServerURL = server
	}
	if apiKey != "" {
		c.APIKey = apiKey
	}
	if output != "" {
		c.Output = output
	}
	if version != "" {
		c.Version = version
	}
}

// Validate checks that required fields are set.
func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("Shlink server URL is required (set SHLINK_SERVER or use --server)")
	}
	if c.APIKey == "" {
		return fmt.Errorf("Shlink API key is required (set SHLINK_API_KEY or use --api-key)")
	}
	c.ServerURL = strings.TrimRight(c.ServerURL, "/")
	return nil
}

// BaseURL returns the versioned API base path.
func (c *Config) BaseURL() string {
	return fmt.Sprintf("%s/rest/v%s", c.ServerURL, c.Version)
}
