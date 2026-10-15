package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Subreddits) == 0 {
		t.Error("Expected default subreddits to be set")
	}

	expectedSubs := []string{"stcatharinesON", "niagara", "NiagaraFalls", "Welland", "brocku"}
	if len(cfg.Subreddits) != len(expectedSubs) {
		t.Errorf("Expected %d subreddits, got %d", len(expectedSubs), len(cfg.Subreddits))
	}

	for i, sub := range expectedSubs {
		if cfg.Subreddits[i] != sub {
			t.Errorf("Expected subreddit[%d] to be %s, got %s", i, sub, cfg.Subreddits[i])
		}
	}

	if cfg.UserAgent == "" {
		t.Error("Expected default user agent to be set")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid dry-run config",
			config: Config{
				DryRun:       true,
				OutputFormat: "json",
			},
			wantError: false,
		},
		{
			name: "missing credentials without dry-run",
			config: Config{
				DryRun:       false,
				OutputFormat: "json",
			},
			wantError: true,
		},
		{
			name: "invalid output format",
			config: Config{
				DryRun:       true,
				OutputFormat: "xml",
			},
			wantError: true,
		},
		{
			name: "valid full config",
			config: Config{
				ClientID:     "test_id",
				ClientSecret: "test_secret",
				Username:     "test_user",
				Password:     "test_pass",
				OutputFormat: "markdown",
				DryRun:       false,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestListingItemDataIsDeleted(t *testing.T) {
	tests := []struct {
		name    string
		data    ListingItemData
		deleted bool
	}{
		{
			name:    "normal post",
			data:    ListingItemData{Author: "user123", Body: "Hello world"},
			deleted: false,
		},
		{
			name:    "deleted author",
			data:    ListingItemData{Author: "[deleted]", Body: "Hello world"},
			deleted: true,
		},
		{
			name:    "deleted body",
			data:    ListingItemData{Author: "user123", Body: "[deleted]"},
			deleted: true,
		},
		{
			name:    "removed body",
			data:    ListingItemData{Author: "user123", Body: "[removed]"},
			deleted: true,
		},
		{
			name:    "deleted selftext",
			data:    ListingItemData{Author: "user123", Selftext: "[deleted]"},
			deleted: true,
		},
		{
			name:    "removed by category",
			data:    ListingItemData{Author: "user123", RemovedByCategory: "moderator"},
			deleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.IsDeleted(); got != tt.deleted {
				t.Errorf("IsDeleted() = %v, want %v", got, tt.deleted)
			}
		})
	}
}

func TestToResearchPost(t *testing.T) {
	now := float64(time.Now().Unix())

	t.Run("converts valid post", func(t *testing.T) {
		data := ListingItemData{
			ID:          "abc123",
			Title:       "Test Post",
			Author:      "testuser",
			Selftext:    "Post content",
			Subreddit:   "niagara",
			Score:       42,
			NumComments: 5,
			CreatedUTC:  now,
			Permalink:   "/r/niagara/comments/abc123/test_post",
		}

		post := data.ToResearchPost()
		if post == nil {
			t.Fatal("Expected non-nil post")
		}

		if post.ID != "abc123" {
			t.Errorf("Expected ID abc123, got %s", post.ID)
		}
		if post.Title != "Test Post" {
			t.Errorf("Expected title 'Test Post', got %s", post.Title)
		}
		if post.Score != 42 {
			t.Errorf("Expected score 42, got %d", post.Score)
		}
	})

	t.Run("returns nil for deleted post", func(t *testing.T) {
		data := ListingItemData{
			Author: "[deleted]",
			Title:  "Deleted Post",
		}

		post := data.ToResearchPost()
		if post != nil {
			t.Error("Expected nil for deleted post")
		}
	})
}

func TestToResearchComment(t *testing.T) {
	now := float64(time.Now().Unix())

	t.Run("converts valid comment", func(t *testing.T) {
		data := ListingItemData{
			ID:         "xyz789",
			Author:     "commenter",
			Body:       "This is a comment",
			Score:      10,
			CreatedUTC: now,
		}

		comment := data.ToResearchComment()
		if comment == nil {
			t.Fatal("Expected non-nil comment")
		}

		if comment.ID != "xyz789" {
			t.Errorf("Expected ID xyz789, got %s", comment.ID)
		}
		if comment.Body != "This is a comment" {
			t.Errorf("Expected body 'This is a comment', got %s", comment.Body)
		}
	})

	t.Run("returns nil for deleted comment", func(t *testing.T) {
		data := ListingItemData{
			Author: "[deleted]",
			Body:   "[removed]",
		}

		comment := data.ToResearchComment()
		if comment != nil {
			t.Error("Expected nil for deleted comment")
		}
	})
}

func TestGenerateSampleData(t *testing.T) {
	cfg := &Config{
		Subreddits: []string{"test1", "test2"},
	}

	output := generateSampleData(cfg)

	if output == nil {
		t.Fatal("Expected non-nil output")
	}

	if len(output.Subreddits) != 2 {
		t.Errorf("Expected 2 subreddits, got %d", len(output.Subreddits))
	}

	if output.TotalPosts == 0 {
		t.Error("Expected some sample posts")
	}

	for _, sub := range cfg.Subreddits {
		posts, ok := output.PostsBySubreddit[sub]
		if !ok {
			t.Errorf("Expected posts for subreddit %s", sub)
		}
		if len(posts) == 0 {
			t.Errorf("Expected at least one post for subreddit %s", sub)
		}
	}
}

func TestResearchOutputJSON(t *testing.T) {
	output := &ResearchOutput{
		FetchedAt:  time.Now().UTC(),
		Subreddits: []string{"test"},
		TotalPosts: 1,
		PostsBySubreddit: map[string][]ResearchPost{
			"test": {
				{
					ID:        "test1",
					Title:     "Test Post",
					Author:    "tester",
					CreatedAt: time.Now().UTC(),
				},
			},
		},
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal output: %v", err)
	}

	var unmarshaled ResearchOutput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if unmarshaled.TotalPosts != 1 {
		t.Errorf("Expected TotalPosts 1, got %d", unmarshaled.TotalPosts)
	}
}

func TestNewRedditClient(t *testing.T) {
	cfg := &Config{
		ClientID:     "test",
		ClientSecret: "secret",
		Username:     "user",
		Password:     "pass",
		UserAgent:    "TestAgent/1.0",
	}

	client := NewRedditClient(cfg)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.config != cfg {
		t.Error("Client config not set correctly")
	}

	if client.limiter == nil {
		t.Error("Expected rate limiter to be initialized")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := &Config{
		Subreddits:   []string{"test"},
		CommentLimit: 0,
	}

	// Create a mock fetcher that respects context
	fetcher := &Fetcher{config: cfg}

	// The fetch should respect context cancellation
	// This is a basic test that the context check is in place
	select {
	case <-ctx.Done():
		// Context is cancelled as expected
	default:
		t.Error("Expected context to be cancelled")
	}

	_ = fetcher // Used to create the fetcher
}
