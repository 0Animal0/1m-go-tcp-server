# Reddit Reader - Niagara Region Research Tool

A read-only Go application for fetching and analyzing public posts and comments from local Niagara region subreddits. This tool is designed for personal, non-commercial research to understand what local residents discuss and ask for help with.

## Purpose

This tool monitors the following Niagara-region subreddits:

- r/stcatharinesON - St. Catharines, Ontario community
- r/niagara - General Niagara region discussions
- r/NiagaraFalls - Niagara Falls community
- r/Welland - Welland community
- r/brocku - Brock University community

The goal is to understand community needs, common questions, and local topics of interest for personal research purposes. This is intended for occasional use (a few times per week) to stay informed about local discussions.

## Features

- **Read-only access**: Only fetches public posts and comments; no posting, voting, or messaging
- **Rate limit compliance**: Respects Reddit's X-Ratelimit headers and implements conservative request throttling
- **Deleted content filtering**: Automatically excludes deleted or removed content
- **Multiple output formats**: JSON for data analysis or Markdown for human-readable summaries
- **Dry-run mode**: Test the tool without credentials using sample data
- **Configurable**: Adjust subreddit list, post limits, and output settings

## Compliance

### Reddit Data API Terms

This tool complies with Reddit's [Data API Terms](https://www.redditinc.com/policies/data-api-terms):

1. **Non-commercial use**: This is personal research software, not for commercial purposes
2. **No data resale**: Collected data is for personal analysis only
3. **Respects rate limits**: Implements proper rate limiting using X-Ratelimit headers
4. **Proper attribution**: Data is from Reddit; this tool does not claim ownership
5. **No deleted content**: Filters out [deleted] and [removed] content

### Responsible Builder Policy

This tool follows Reddit's [Responsible Builder Policy](https://support.reddithelp.com/hc/en-us/articles/26410253549972-Reddit-Data-API-Wiki):

1. **Descriptive User-Agent**: Uses a clear, descriptive user agent string identifying the application
2. **OAuth authentication**: Uses the official OAuth2 script app flow
3. **Low volume**: Designed for occasional use, not bulk data collection
4. **Read-only scope**: Only requests read permissions, never write access
5. **Respectful of users**: Does not track or store user data beyond post/comment metadata

## Installation

### Prerequisites

- Go 1.21 or later
- Reddit API credentials (see Setup section)

### Build

```bash
cd reddit-reader
go mod tidy
go build -o reddit-reader .
```

## Setup

### 1. Create a Reddit Application

1. Go to https://www.reddit.com/prefs/apps
2. Click "create another app..."
3. Select "script" as the application type
4. Fill in the required fields:
   - **name**: NiagaraLocalResearch (or similar)
   - **description**: Personal research tool for understanding local Niagara community discussions
   - **redirect uri**: http://localhost:8080 (not used for script apps, but required)
5. Click "create app"
6. Note your client ID (under the app name) and client secret

### 2. Set Environment Variables

```bash
export REDDIT_CLIENT_ID="your_client_id"
export REDDIT_CLIENT_SECRET="your_client_secret"
export REDDIT_USERNAME="your_reddit_username"
export REDDIT_PASSWORD="your_reddit_password"

# Optional: Custom user agent
export REDDIT_USER_AGENT="NiagaraLocalResearchBot/1.0 (personal research; non-commercial)"

# Optional: Override subreddit list
export REDDIT_SUBREDDITS="stcatharinesON,niagara,NiagaraFalls"

# Optional: Output format (json or markdown)
export REDDIT_OUTPUT_FORMAT="markdown"
```

**Important**: Never commit credentials to version control. Use environment variables or a `.env` file (excluded from git).

## Usage

### Basic Usage

```bash
# Run with default settings (JSON output)
./reddit-reader

# Output as Markdown
./reddit-reader -format markdown

# Custom output filename
./reddit-reader -output niagara_weekly_summary

# Limit posts and comments
./reddit-reader -limit 10 -comments 5
```

### Dry-Run Mode

Test the tool without credentials:

```bash
./reddit-reader -dry-run
./reddit-reader -dry-run -format markdown
```

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-dry-run` | Run with sample data (no credentials needed) | `false` |
| `-format` | Output format: `json` or `markdown` | `json` |
| `-output` | Output filename (without extension) | `reddit_research_output` |
| `-limit` | Maximum posts per subreddit per sort | `25` |
| `-comments` | Maximum top-level comments per post | `10` |

### Running Tests

```bash
go test -v ./...
```

## Output

### JSON Format

```json
{
  "fetched_at": "2024-01-15T10:30:00Z",
  "subreddits": ["stcatharinesON", "niagara", "NiagaraFalls", "Welland", "brocku"],
  "total_posts": 50,
  "posts_by_subreddit": {
    "stcatharinesON": [
      {
        "id": "abc123",
        "subreddit": "stcatharinesON",
        "title": "Best local restaurants?",
        "author": "local_foodie",
        "content": "Looking for recommendations...",
        "score": 25,
        "num_comments": 15,
        "created_at": "2024-01-14T18:00:00Z",
        "top_comments": [...]
      }
    ]
  }
}
```

### Markdown Format

The Markdown output provides a human-readable summary suitable for quick review.

## Rate Limiting

This tool implements conservative rate limiting:

1. **Client-side throttling**: Maximum 1 request per 2 seconds
2. **Server-side compliance**: Reads and respects Reddit's X-Ratelimit headers
3. **Backoff on limit**: Waits when X-Ratelimit-Remaining approaches zero
4. **429 handling**: Respects Retry-After header on rate limit responses

## Data Privacy

- Only public posts and comments are fetched
- No private messages or user profiles are accessed
- Deleted/removed content is filtered out
- Data is stored locally only
- No data is shared with third parties

## License

This tool is for personal use. The code is provided as-is for the purpose of demonstrating compliance with Reddit's API terms.

## Disclaimer

This tool is not affiliated with, endorsed by, or officially connected with Reddit, Inc. Reddit is a registered trademark of Reddit, Inc.

The data accessed through this tool remains the property of Reddit and its users. Use this tool responsibly and in accordance with Reddit's terms of service.
