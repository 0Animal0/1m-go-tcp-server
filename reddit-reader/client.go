package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	tokenURL = "https://www.reddit.com/api/v1/access_token"
	apiBase  = "https://oauth.reddit.com"
)

// RedditClient handles authenticated requests to the Reddit API.
type RedditClient struct {
	httpClient  *http.Client
	config      *Config
	accessToken string
	tokenExpiry time.Time
	mu          sync.Mutex
	limiter     *rate.Limiter

	// Rate limit info from Reddit headers
	remaining   float64
	resetTime   time.Time
	rateLimitMu sync.Mutex
}

// NewRedditClient creates a new Reddit API client.
func NewRedditClient(cfg *Config) *RedditClient {
	return &RedditClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     cfg,
		limiter:    rate.NewLimiter(rate.Every(2*time.Second), 1), // Conservative: 1 request per 2 seconds
		remaining:  60,
	}
}

// authenticate obtains an OAuth2 access token using the script app flow.
func (c *RedditClient) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if current token is still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-time.Minute)) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.config.Username)
	data.Set("password", c.config.Password)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}

	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// doRequest performs an authenticated API request with rate limiting.
func (c *RedditClient) doRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, err
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	// Check Reddit's rate limit headers
	c.rateLimitMu.Lock()
	if c.remaining <= 1 && time.Now().Before(c.resetTime) {
		waitDuration := time.Until(c.resetTime) + time.Second
		c.rateLimitMu.Unlock()
		fmt.Printf("Rate limit reached, waiting %v...\n", waitDuration)
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	} else {
		c.rateLimitMu.Unlock()
	}

	reqURL := apiBase + endpoint
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	c.mu.Lock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.Unlock()
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse rate limit headers
	c.updateRateLimits(resp.Header)

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		waitSec, _ := strconv.Atoi(retryAfter)
		if waitSec == 0 {
			waitSec = 60
		}
		return nil, fmt.Errorf("rate limited, retry after %d seconds", waitSec)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// updateRateLimits parses Reddit's rate limit headers.
func (c *RedditClient) updateRateLimits(headers http.Header) {
	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	if remaining := headers.Get("X-Ratelimit-Remaining"); remaining != "" {
		if val, err := strconv.ParseFloat(remaining, 64); err == nil {
			c.remaining = val
		}
	}

	if reset := headers.Get("X-Ratelimit-Reset"); reset != "" {
		if val, err := strconv.Atoi(reset); err == nil {
			c.resetTime = time.Now().Add(time.Duration(val) * time.Second)
		}
	}
}

// GetSubredditPosts fetches recent posts from a subreddit.
func (c *RedditClient) GetSubredditPosts(ctx context.Context, subreddit string, sort string, limit int) (*SubredditListing, error) {
	endpoint := fmt.Sprintf("/r/%s/%s.json?limit=%d&raw_json=1", subreddit, sort, limit)

	data, err := c.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var listing SubredditListing
	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, fmt.Errorf("parsing subreddit listing: %w", err)
	}

	return &listing, nil
}

// GetPostComments fetches top-level comments for a post.
func (c *RedditClient) GetPostComments(ctx context.Context, subreddit, postID string, limit int) (*CommentsResponse, error) {
	endpoint := fmt.Sprintf("/r/%s/comments/%s.json?limit=%d&depth=1&raw_json=1", subreddit, postID, limit)

	data, err := c.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var listings []json.RawMessage
	if err := json.Unmarshal(data, &listings); err != nil {
		return nil, fmt.Errorf("parsing comments response: %w", err)
	}

	if len(listings) < 2 {
		return &CommentsResponse{}, nil
	}

	var commentsListing SubredditListing
	if err := json.Unmarshal(listings[1], &commentsListing); err != nil {
		return nil, fmt.Errorf("parsing comments listing: %w", err)
	}

	return &CommentsResponse{Comments: commentsListing}, nil
}
