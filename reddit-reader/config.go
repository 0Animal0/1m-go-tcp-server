package main

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	UserAgent    string
	Subreddits   []string
	OutputFormat string // "json" or "markdown"
	OutputFile   string
	PostLimit    int
	CommentLimit int
	DryRun       bool
}

// DefaultSubreddits is the list of Niagara-region subreddits to monitor.
var DefaultSubreddits = []string{
	"stcatharinesON",
	"niagara",
	"NiagaraFalls",
	"Welland",
	"brocku",
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ClientID:     os.Getenv("REDDIT_CLIENT_ID"),
		ClientSecret: os.Getenv("REDDIT_CLIENT_SECRET"),
		Username:     os.Getenv("REDDIT_USERNAME"),
		Password:     os.Getenv("REDDIT_PASSWORD"),
		UserAgent:    os.Getenv("REDDIT_USER_AGENT"),
		Subreddits:   DefaultSubreddits,
		OutputFormat: "json",
		OutputFile:   "reddit_research_output",
		PostLimit:    25,
		CommentLimit: 10,
		DryRun:       false,
	}

	// Allow custom subreddit list via environment
	if subs := os.Getenv("REDDIT_SUBREDDITS"); subs != "" {
		cfg.Subreddits = strings.Split(subs, ",")
		for i := range cfg.Subreddits {
			cfg.Subreddits[i] = strings.TrimSpace(cfg.Subreddits[i])
		}
	}

	// Allow output format override
	if format := os.Getenv("REDDIT_OUTPUT_FORMAT"); format != "" {
		cfg.OutputFormat = strings.ToLower(format)
	}

	// Default user agent if not provided
	if cfg.UserAgent == "" {
		cfg.UserAgent = "NiagaraLocalResearchBot/1.0 (personal research; non-commercial)"
	}

	return cfg, nil
}

// Validate checks that required credentials are present (unless in dry-run mode).
func (c *Config) Validate() error {
	// Always validate output format
	if c.OutputFormat != "json" && c.OutputFormat != "markdown" {
		return fmt.Errorf("output format must be 'json' or 'markdown'")
	}

	if c.DryRun {
		return nil
	}

	if c.ClientID == "" {
		return fmt.Errorf("REDDIT_CLIENT_ID environment variable is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("REDDIT_CLIENT_SECRET environment variable is required")
	}
	if c.Username == "" {
		return fmt.Errorf("REDDIT_USERNAME environment variable is required")
	}
	if c.Password == "" {
		return fmt.Errorf("REDDIT_PASSWORD environment variable is required")
	}

	return nil
}
