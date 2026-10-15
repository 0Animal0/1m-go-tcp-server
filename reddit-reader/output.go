package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// OutputWriter handles writing research output to files.
type OutputWriter struct {
	config *Config
}

// NewOutputWriter creates a new output writer.
func NewOutputWriter(cfg *Config) *OutputWriter {
	return &OutputWriter{config: cfg}
}

// Write outputs the research data in the configured format.
func (w *OutputWriter) Write(output *ResearchOutput) error {
	switch w.config.OutputFormat {
	case "json":
		return w.writeJSON(output)
	case "markdown":
		return w.writeMarkdown(output)
	default:
		return fmt.Errorf("unsupported output format: %s", w.config.OutputFormat)
	}
}

func (w *OutputWriter) writeJSON(output *ResearchOutput) error {
	filename := w.config.OutputFile + ".json"
	
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Output written to %s\n", filename)
	return nil
}

func (w *OutputWriter) writeMarkdown(output *ResearchOutput) error {
	filename := w.config.OutputFile + ".md"

	var sb strings.Builder

	sb.WriteString("# Niagara Region Reddit Research Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Fetched:** %s\n\n", output.FetchedAt.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("**Subreddits:** %s\n\n", strings.Join(output.Subreddits, ", ")))
	sb.WriteString(fmt.Sprintf("**Total Posts:** %d\n\n", output.TotalPosts))
	sb.WriteString("---\n\n")

	for _, subreddit := range output.Subreddits {
		posts, ok := output.PostsBySubreddit[subreddit]
		if !ok || len(posts) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## r/%s\n\n", subreddit))

		for _, post := range posts {
			sb.WriteString(fmt.Sprintf("### %s\n\n", post.Title))
			sb.WriteString(fmt.Sprintf("- **Author:** u/%s\n", post.Author))
			sb.WriteString(fmt.Sprintf("- **Score:** %d | **Comments:** %d\n", post.Score, post.NumComments))
			sb.WriteString(fmt.Sprintf("- **Posted:** %s\n", post.CreatedAt.Format("Jan 2, 2006 3:04 PM")))
			if post.Flair != "" {
				sb.WriteString(fmt.Sprintf("- **Flair:** %s\n", post.Flair))
			}
			sb.WriteString(fmt.Sprintf("- **Link:** [View Post](%s)\n", post.Permalink))
			sb.WriteString("\n")

			if post.Content != "" {
				content := post.Content
				if len(content) > 500 {
					content = content[:500] + "..."
				}
				sb.WriteString(fmt.Sprintf("> %s\n\n", strings.ReplaceAll(content, "\n", "\n> ")))
			}

			if len(post.TopComments) > 0 {
				sb.WriteString("**Top Comments:**\n\n")
				for _, comment := range post.TopComments {
					body := comment.Body
					if len(body) > 300 {
						body = body[:300] + "..."
					}
					sb.WriteString(fmt.Sprintf("- **u/%s** (score: %d): %s\n\n",
						comment.Author, comment.Score, strings.ReplaceAll(body, "\n", " ")))
				}
			}

			sb.WriteString("---\n\n")
		}
	}

	sb.WriteString("\n*This data was collected for personal, non-commercial research purposes ")
	sb.WriteString("to understand community needs in the Niagara region.*\n")

	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Output written to %s\n", filename)
	return nil
}
