package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/bigwhite/issue2md/internal/github"
)

func ptr[T any](v T) *T { return &v }

func TestConvertIssue(t *testing.T) {
	now := time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name   string
		issue  *github.Issue
		opts   Option
		checks []string // must all appear
		absent []string // must NOT appear
	}{
		{
			name: "basic issue with all fields",
			issue: &github.Issue{
				GitResource: github.GitHubResource{
					Number:    42,
					Title:     "Fix login timeout issue",
					URL:       "https://github.com/owner/repo/issues/42",
					Author:    "octocat",
					State:     "open",
					Body:      "This is the issue description.",
					CreatedAt: now,
					UpdatedAt: now,
					Labels:    []string{"bug", "performance"},
					Milestone: "v2.0",
				},
				Comments: []github.Comment{
					{
						Author:    "user1",
						Body:      "I can reproduce this.",
						CreatedAt: now.Add(1 * time.Hour),
						UpdatedAt: now.Add(1 * time.Hour),
					},
				},
			},
			opts: Option{},
			checks: []string{
				"title: \"Fix login timeout issue\"",
				"url: https://github.com/owner/repo/issues/42",
				"number: 42",
				"type: issue",
				"author: octocat",
				"state: open",
				"created_at: \"2026-04-15T10:30:00Z\"",
				"  - bug",
				"  - performance",
				"milestone: v2.0",
				"# Fix login timeout issue (#42)",
				"This is the issue description.",
				"## Comments",
				"[Comment] by **user1**",
				"I can reproduce this.",
			},
			absent: []string{
				"Reactions",
				"github.com/octocat",
			},
		},
		{
			name: "issue with reactions enabled",
			issue: &github.Issue{
				GitResource: github.GitHubResource{
					Number:    7,
					Title:     "Add dark mode",
					URL:       "https://github.com/owner/repo/issues/7",
					Author:    "dev1",
					State:     "open",
					Body:      "Body text",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "user2",
						Body:      "Great idea!",
						CreatedAt: now.Add(1 * time.Hour),
						UpdatedAt: now.Add(1 * time.Hour),
						Reactions: github.ReactionCount{
							ThumbsUp: 5,
							Heart:    2,
						},
					},
				},
			},
			opts: Option{EnableReactions: true},
			checks: []string{
				"type: issue",
				"# Add dark mode (#7)",
				"Great idea!",
				"Reactions: 👍 5 ❤️ 2",
			},
		},
		{
			name: "issue with user links enabled",
			issue: &github.Issue{
				GitResource: github.GitHubResource{
					Number:    10,
					Title:     "Mention test",
					URL:       "https://github.com/owner/repo/issues/10",
					Author:    "leader",
					State:     "open",
					Body:      "CC @user1 @user2 please review",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "user1",
						Body:      "Thanks @leader, will do",
						CreatedAt: now.Add(1 * time.Hour),
						UpdatedAt: now.Add(1 * time.Hour),
					},
				},
			},
			opts: Option{EnableUserLinks: true},
			checks: []string{
				"CC [@user1](https://github.com/user1) [@user2](https://github.com/user2)",
				"Thanks [@leader](https://github.com/leader)",
			},
			absent: []string{
				"CC @user1",
				"Thanks @leader",
			},
		},
		{
			name: "issue with no comments and no labels",
			issue: &github.Issue{
				GitResource: github.GitHubResource{
					Number:    1,
					Title:     "Minimal",
					URL:       "https://github.com/a/b/issues/1",
					Author:    "bot",
					State:     "closed",
					Body:      "Fixed.",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			opts: Option{},
			checks: []string{
				"author: bot",
				"state: closed",
				"labels: []",
				"milestone: \"\"",
				"# Minimal (#1)",
				"Fixed.",
			},
			absent: []string{
				"## Comments",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.opts)
			got, err := c.ConvertIssue(tt.issue)
			if err != nil {
				t.Fatalf("ConvertIssue() returned error: %v", err)
			}
			for _, ch := range tt.checks {
				if !strings.Contains(got, ch) {
					t.Errorf("expected output to contain:\n  %q\ngot:\n%s", ch, got)
				}
			}
			for _, ab := range tt.absent {
				if strings.Contains(got, ab) {
					t.Errorf("expected output NOT to contain:\n  %q\ngot:\n%s", ab, got)
				}
			}
		})
	}
}

func TestConvertPullRequest(t *testing.T) {
	now := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		pr     *github.PullRequest
		opts   Option
		checks []string
		absent []string
	}{
		{
			name: "PR with issue comment and review comment",
			pr: &github.PullRequest{
				GitResource: github.GitHubResource{
					Number:    500,
					Title:     "Fix race condition in http.Server",
					URL:       "https://github.com/golang/go/pull/500",
					Author:    "bradfitz",
					State:     "open",
					Body:      "This PR fixes a race condition.",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "iansmith",
						Body:      "Have you considered the timeout case?",
						CreatedAt: now.Add(3 * time.Hour),
						UpdatedAt: now.Add(3 * time.Hour),
					},
					{
						Author:    "dvyukov",
						Body:      "There's a race on this line.",
						CreatedAt: now.Add(2 * time.Hour),
						UpdatedAt: now.Add(2 * time.Hour),
						FilePath:  "net/http/server.go",
						Line:      342,
					},
				},
			},
			opts: Option{},
			checks: []string{
				"type: pull",
				"author: bradfitz",
				"# Fix race condition in http.Server (#500)",
				"This PR fixes a race condition.",
				"[Review Comment] by **dvyukov** on `net/http/server.go` (line 342)",
				"[Comment] by **iansmith**",
				"Have you considered the timeout case?",
				"There's a race on this line.",
			},
		},
		{
			name: "PR comments sorted by time",
			pr: &github.PullRequest{
				GitResource: github.GitHubResource{
					Number:    1,
					Title:     "Sorted comments",
					URL:       "https://github.com/a/b/pull/1",
					Author:    "user1",
					State:     "open",
					Body:      "PR body.",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "late",
						Body:      "Third comment",
						CreatedAt: now.Add(3 * time.Hour),
					},
					{
						Author:    "early",
						Body:      "First comment",
						CreatedAt: now.Add(1 * time.Hour),
					},
					{
						Author:    "middle",
						Body:      "Second comment",
						CreatedAt: now.Add(2 * time.Hour),
					},
				},
			},
			opts: Option{},
			checks: []string{
				"First comment",
				"Second comment",
				"Third comment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.opts)
			got, err := c.ConvertPullRequest(tt.pr)
			if err != nil {
				t.Fatalf("ConvertPullRequest() returned error: %v", err)
			}
			for _, ch := range tt.checks {
				if !strings.Contains(got, ch) {
					t.Errorf("expected output to contain:\n  %q\ngot:\n%s", ch, got)
				}
			}
			for _, ab := range tt.absent {
				if strings.Contains(got, ab) {
					t.Errorf("expected output NOT to contain:\n  %q\ngot:\n%s", ab, got)
				}
			}
		})
	}
}

func TestConvertDiscussion(t *testing.T) {
	now := time.Date(2026, 4, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		disc   *github.Discussion
		opts   Option
		checks []string
		absent []string
	}{
		{
			name: "discussion with answer",
			disc: &github.Discussion{
				GitResource: github.GitHubResource{
					Number:    42,
					Title:     "Best practices for error handling",
					URL:       "https://github.com/community/community/discussions/42",
					Author:    "gopher",
					State:     "open",
					Body:      "What are the best practices?",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "davecheney",
						Body:      "Use fmt.Errorf with %w for wrapping.",
						CreatedAt: now.Add(24 * time.Hour),
						UpdatedAt: now.Add(24 * time.Hour),
						IsAnswer:  true,
					},
					{
						Author:    "robpike",
						Body:      "I agree with Dave.",
						CreatedAt: now.Add(48 * time.Hour),
						UpdatedAt: now.Add(48 * time.Hour),
					},
				},
			},
			opts: Option{},
			checks: []string{
				"type: discussion",
				"# Best practices for error handling (#42)",
				"What are the best practices?",
				"✅ Answer by **davecheney**",
				"Use fmt.Errorf with %w for wrapping.",
				"[Comment] by **robpike**",
				"I agree with Dave.",
			},
		},
		{
			name: "discussion with nested replies",
			disc: &github.Discussion{
				GitResource: github.GitHubResource{
					Number:    99,
					Title:     "Nested thread",
					URL:       "https://github.com/org/repo/discussions/99",
					Author:    "opener",
					State:     "open",
					Body:      "Original post.",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []github.Comment{
					{
						Author:    "parent",
						Body:      "Top level comment.",
						CreatedAt: now.Add(1 * time.Hour),
					},
					{
						Author:    "child",
						Body:      "Reply to parent.",
						CreatedAt: now.Add(2 * time.Hour),
						ParentID:  ptr(int64(1)),
					},
				},
			},
			opts: Option{},
			checks: []string{
				"Top level comment.",
				"Reply to parent.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.opts)
			got, err := c.ConvertDiscussion(tt.disc)
			if err != nil {
				t.Fatalf("ConvertDiscussion() returned error: %v", err)
			}
			for _, ch := range tt.checks {
				if !strings.Contains(got, ch) {
					t.Errorf("expected output to contain:\n  %q\ngot:\n%s", ch, got)
				}
			}
			for _, ab := range tt.absent {
				if strings.Contains(got, ab) {
					t.Errorf("expected output NOT to contain:\n  %q\ngot:\n%s", ab, got)
				}
			}
		})
	}
}

func TestNewConverter(t *testing.T) {
	c := New(Option{EnableReactions: true, EnableUserLinks: true})
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestConvertIssue_NilInput(t *testing.T) {
	c := New(Option{})
	_, err := c.ConvertIssue(nil)
	if err == nil {
		t.Error("expected error for nil issue, got nil")
	}
}

func TestConvertPullRequest_NilInput(t *testing.T) {
	c := New(Option{})
	_, err := c.ConvertPullRequest(nil)
	if err == nil {
		t.Error("expected error for nil PR, got nil")
	}
}

func TestConvertDiscussion_NilInput(t *testing.T) {
	c := New(Option{})
	_, err := c.ConvertDiscussion(nil)
	if err == nil {
		t.Error("expected error for nil discussion, got nil")
	}
}
