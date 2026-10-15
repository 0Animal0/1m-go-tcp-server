package main

import (
	"context"
	"fmt"
	"time"
)

// Fetcher coordinates fetching posts and comments from multiple subreddits.
type Fetcher struct {
	client *RedditClient
	config *Config
}

// NewFetcher creates a new fetcher.
func NewFetcher(client *RedditClient, cfg *Config) *Fetcher {
	return &Fetcher{
		client: client,
		config: cfg,
	}
}

// FetchAll retrieves posts and comments from all configured subreddits.
func (f *Fetcher) FetchAll(ctx context.Context) (*ResearchOutput, error) {
	output := &ResearchOutput{
		FetchedAt:        time.Now().UTC(),
		Subreddits:       f.config.Subreddits,
		PostsBySubreddit: make(map[string][]ResearchPost),
	}

	for _, subreddit := range f.config.Subreddits {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fmt.Printf("Fetching posts from r/%s...\n", subreddit)

		posts, err := f.fetchSubredditPosts(ctx, subreddit)
		if err != nil {
			fmt.Printf("Warning: failed to fetch r/%s: %v\n", subreddit, err)
			continue
		}

		if len(posts) > 0 {
			output.PostsBySubreddit[subreddit] = posts
			output.TotalPosts += len(posts)
		}

		fmt.Printf("  Found %d posts\n", len(posts))
	}

	return output, nil
}

func (f *Fetcher) fetchSubredditPosts(ctx context.Context, subreddit string) ([]ResearchPost, error) {
	var allPosts []ResearchPost
	seenIDs := make(map[string]bool)

	// Fetch from both "new" and "top" (week) to get a mix of recent and popular
	for _, sort := range []string{"new", "top"} {
		listing, err := f.client.GetSubredditPosts(ctx, subreddit, sort, f.config.PostLimit)
		if err != nil {
			return nil, fmt.Errorf("fetching %s posts: %w", sort, err)
		}

		for _, child := range listing.Data.Children {
			if child.Kind != "t3" { // t3 = post/link
				continue
			}

			// Skip deleted content
			if child.Data.IsDeleted() {
				continue
			}

			// Skip duplicates
			if seenIDs[child.Data.ID] {
				continue
			}
			seenIDs[child.Data.ID] = true

			post := child.Data.ToResearchPost()
			if post == nil {
				continue
			}

			// Fetch top comments for posts with comments
			if child.Data.NumComments > 0 && f.config.CommentLimit > 0 {
				comments, err := f.fetchPostComments(ctx, subreddit, child.Data.ID)
				if err != nil {
					fmt.Printf("  Warning: couldn't fetch comments for %s: %v\n", child.Data.ID, err)
				} else {
					post.TopComments = comments
				}
			}

			allPosts = append(allPosts, *post)
		}
	}

	return allPosts, nil
}

func (f *Fetcher) fetchPostComments(ctx context.Context, subreddit, postID string) ([]ResearchComment, error) {
	commentsResp, err := f.client.GetPostComments(ctx, subreddit, postID, f.config.CommentLimit)
	if err != nil {
		return nil, err
	}

	var comments []ResearchComment
	for _, child := range commentsResp.Comments.Data.Children {
		if child.Kind != "t1" { // t1 = comment
			continue
		}

		// Skip deleted content
		if child.Data.IsDeleted() {
			continue
		}

		comment := child.Data.ToResearchComment()
		if comment != nil {
			comments = append(comments, *comment)
		}

		if len(comments) >= f.config.CommentLimit {
			break
		}
	}

	return comments, nil
}
