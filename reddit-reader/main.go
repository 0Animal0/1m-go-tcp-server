// Reddit Reader - A read-only tool for researching local Niagara region subreddits
//
// This tool fetches public posts and comments from specified subreddits for
// personal, non-commercial research purposes. It complies with Reddit's
// Data API Terms and Responsible Builder Policy.
//
// Usage:
//
//	go run . [flags]
//
// Flags:
//
//	-dry-run     Run in test mode without credentials (uses sample data)
//	-format      Output format: "json" or "markdown" (default: json)
//	-output      Output filename without extension (default: reddit_research_output)
//	-limit       Maximum posts per subreddit sort (default: 25)
//	-comments    Maximum comments per post (default: 10)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Parse command-line flags
	dryRun := flag.Bool("dry-run", false, "Run in test mode without credentials")
	format := flag.String("format", "", "Output format: json or markdown")
	output := flag.String("output", "", "Output filename without extension")
	limit := flag.Int("limit", 0, "Maximum posts per subreddit sort")
	comments := flag.Int("comments", 0, "Maximum comments per post")
	flag.Parse()

	// Load and configure
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply flag overrides
	cfg.DryRun = *dryRun
	if *format != "" {
		cfg.OutputFormat = *format
	}
	if *output != "" {
		cfg.OutputFile = *output
	}
	if *limit > 0 {
		cfg.PostLimit = *limit
	}
	if *comments >= 0 {
		cfg.CommentLimit = *comments
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nRequired environment variables:")
		fmt.Fprintln(os.Stderr, "  REDDIT_CLIENT_ID     - Reddit API client ID")
		fmt.Fprintln(os.Stderr, "  REDDIT_CLIENT_SECRET - Reddit API client secret")
		fmt.Fprintln(os.Stderr, "  REDDIT_USERNAME      - Reddit username")
		fmt.Fprintln(os.Stderr, "  REDDIT_PASSWORD      - Reddit password")
		fmt.Fprintln(os.Stderr, "\nOptional environment variables:")
		fmt.Fprintln(os.Stderr, "  REDDIT_USER_AGENT    - Custom user agent string")
		fmt.Fprintln(os.Stderr, "  REDDIT_SUBREDDITS    - Comma-separated subreddit list")
		fmt.Fprintln(os.Stderr, "  REDDIT_OUTPUT_FORMAT - Output format (json/markdown)")
		fmt.Fprintln(os.Stderr, "\nOr use -dry-run flag to test without credentials")
		os.Exit(1)
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	// Run the appropriate mode
	var output_data *ResearchOutput

	if cfg.DryRun {
		fmt.Println("Running in dry-run mode with sample data...")
		output_data = generateSampleData(cfg)
	} else {
		fmt.Printf("Fetching data from %d subreddits...\n", len(cfg.Subreddits))
		fmt.Printf("User-Agent: %s\n\n", cfg.UserAgent)

		client := NewRedditClient(cfg)
		fetcher := NewFetcher(client, cfg)

		output_data, err = fetcher.FetchAll(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching data: %v\n", err)
			os.Exit(1)
		}
	}

	// Write output
	writer := NewOutputWriter(cfg)
	if err := writer.Write(output_data); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone! Fetched %d total posts from %d subreddits.\n",
		output_data.TotalPosts, len(output_data.Subreddits))
}

// generateSampleData creates sample output for dry-run testing.
func generateSampleData(cfg *Config) *ResearchOutput {
	output := &ResearchOutput{
		FetchedAt:        time.Now().UTC(),
		Subreddits:       cfg.Subreddits,
		PostsBySubreddit: make(map[string][]ResearchPost),
	}

	for _, sub := range cfg.Subreddits {
		posts := []ResearchPost{
			{
				ID:          "sample1",
				Subreddit:   sub,
				Title:       "[DRY-RUN] Sample post about local services",
				Author:      "sample_user",
				Content:     "This is sample content for testing. In a real run, this would contain actual post text.",
				URL:         "https://reddit.com/r/" + sub,
				Permalink:   "https://reddit.com/r/" + sub + "/comments/sample1",
				Score:       42,
				NumComments: 5,
				CreatedAt:   time.Now().Add(-24 * time.Hour).UTC(),
				Flair:       "Question",
				TopComments: []ResearchComment{
					{
						ID:        "comment1",
						Author:    "helpful_neighbor",
						Body:      "This is a sample comment for testing purposes.",
						Score:     10,
						CreatedAt: time.Now().Add(-23 * time.Hour).UTC(),
					},
				},
			},
			{
				ID:          "sample2",
				Subreddit:   sub,
				Title:       "[DRY-RUN] Another sample community discussion",
				Author:      "local_resident",
				Content:     "Sample discussion content for the dry-run test mode.",
				URL:         "https://reddit.com/r/" + sub,
				Permalink:   "https://reddit.com/r/" + sub + "/comments/sample2",
				Score:       28,
				NumComments: 3,
				CreatedAt:   time.Now().Add(-48 * time.Hour).UTC(),
			},
		}
		output.PostsBySubreddit[sub] = posts
		output.TotalPosts += len(posts)
	}

	return output
}
