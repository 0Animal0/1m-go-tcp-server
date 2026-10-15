package main

import "time"

// SubredditListing represents the response from Reddit's listing endpoints.
type SubredditListing struct {
	Kind string `json:"kind"`
	Data struct {
		After    string       `json:"after"`
		Before   string       `json:"before"`
		Children []ListingChild `json:"children"`
	} `json:"data"`
}

// ListingChild represents an item in a Reddit listing.
type ListingChild struct {
	Kind string          `json:"kind"`
	Data ListingItemData `json:"data"`
}

// ListingItemData represents the data for a post or comment.
type ListingItemData struct {
	// Common fields
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Author    string  `json:"author"`
	Body      string  `json:"body,omitempty"`
	CreatedUTC float64 `json:"created_utc"`
	Score     int     `json:"score"`
	Permalink string  `json:"permalink"`

	// Post-specific fields
	Title         string `json:"title,omitempty"`
	Selftext      string `json:"selftext,omitempty"`
	URL           string `json:"url,omitempty"`
	Subreddit     string `json:"subreddit"`
	NumComments   int    `json:"num_comments,omitempty"`
	IsVideo       bool   `json:"is_video,omitempty"`
	IsSelf        bool   `json:"is_self,omitempty"`
	LinkFlairText string `json:"link_flair_text,omitempty"`
	Upvotes       int    `json:"ups"`
	Downvotes     int    `json:"downs"`

	// Deleted/removed content indicators
	RemovedByCategory string `json:"removed_by_category,omitempty"`
}

// CommentsResponse holds comment data for a post.
type CommentsResponse struct {
	Comments SubredditListing
}

// ResearchPost represents a processed post for output.
type ResearchPost struct {
	ID            string            `json:"id"`
	Subreddit     string            `json:"subreddit"`
	Title         string            `json:"title"`
	Author        string            `json:"author"`
	Content       string            `json:"content,omitempty"`
	URL           string            `json:"url"`
	Permalink     string            `json:"permalink"`
	Score         int               `json:"score"`
	NumComments   int               `json:"num_comments"`
	CreatedAt     time.Time         `json:"created_at"`
	Flair         string            `json:"flair,omitempty"`
	TopComments   []ResearchComment `json:"top_comments,omitempty"`
}

// ResearchComment represents a processed comment for output.
type ResearchComment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// ResearchOutput represents the complete research output.
type ResearchOutput struct {
	FetchedAt   time.Time               `json:"fetched_at"`
	Subreddits  []string                `json:"subreddits"`
	TotalPosts  int                     `json:"total_posts"`
	PostsBySubreddit map[string][]ResearchPost `json:"posts_by_subreddit"`
}

// IsDeleted checks if content has been deleted or removed.
func (d *ListingItemData) IsDeleted() bool {
	return d.Author == "[deleted]" ||
		d.Body == "[deleted]" ||
		d.Body == "[removed]" ||
		d.Selftext == "[deleted]" ||
		d.Selftext == "[removed]" ||
		d.RemovedByCategory != ""
}

// ToResearchPost converts API data to a ResearchPost.
func (d *ListingItemData) ToResearchPost() *ResearchPost {
	if d.IsDeleted() {
		return nil
	}

	return &ResearchPost{
		ID:          d.ID,
		Subreddit:   d.Subreddit,
		Title:       d.Title,
		Author:      d.Author,
		Content:     d.Selftext,
		URL:         d.URL,
		Permalink:   "https://reddit.com" + d.Permalink,
		Score:       d.Score,
		NumComments: d.NumComments,
		CreatedAt:   time.Unix(int64(d.CreatedUTC), 0).UTC(),
		Flair:       d.LinkFlairText,
	}
}

// ToResearchComment converts API data to a ResearchComment.
func (d *ListingItemData) ToResearchComment() *ResearchComment {
	if d.IsDeleted() {
		return nil
	}

	return &ResearchComment{
		ID:        d.ID,
		Author:    d.Author,
		Body:      d.Body,
		Score:     d.Score,
		CreatedAt: time.Unix(int64(d.CreatedUTC), 0).UTC(),
	}
}
