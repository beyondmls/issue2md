package github

import (
	"context"
	"testing"
)

func TestFetchDiscussion_Real(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	// community/community is a well-known public discussions repository
	disc, err := client.FetchDiscussion(context.Background(), "community", "community", 1)
	if err != nil {
		t.Fatalf("FetchDiscussion failed: %v", err)
	}

	if disc.GitResource.Number != 1 {
		t.Errorf("expected number 1, got %d", disc.GitResource.Number)
	}
	if disc.GitResource.Author == "" {
		t.Error("expected non-empty author")
	}
	if disc.GitResource.Title == "" {
		t.Error("expected non-empty title")
	}
	if disc.GitResource.State == "" {
		t.Error("expected non-empty state")
	}

	t.Logf("Discussion #1: %q by %s (%s)", disc.GitResource.Title, disc.GitResource.Author, disc.GitResource.State)
	t.Logf("  comments: %d", len(disc.Comments))

	// Verify comments are sorted by CreatedAt
	for i := 1; i < len(disc.Comments); i++ {
		if disc.Comments[i].CreatedAt.Before(disc.Comments[i-1].CreatedAt) {
			t.Errorf("comments not sorted by CreatedAt: index %d (%v) before %d (%v)",
				i-1, disc.Comments[i-1].CreatedAt, i, disc.Comments[i].CreatedAt)
		}
	}
}

func TestFetchDiscussion_NotFound(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	_, err := client.FetchDiscussion(context.Background(), "golang", "go", 99999999)
	if err == nil {
		t.Fatal("expected error for non-existent discussion, got nil")
	}
	t.Logf("Got expected error: %v", err)
}

func TestFetchDiscussion_NoToken(t *testing.T) {
	client := NewClient("")
	_, err := client.FetchDiscussion(context.Background(), "community", "community", 1)
	if err == nil {
		t.Fatal("expected error when GITHUB_TOKEN is not set for GraphQL")
	}
	t.Logf("Got expected error: %v", err)
}

func TestParseDiscussionResponse_Valid(t *testing.T) {
	// Test with minimum valid JSON
	body := []byte(`{
		"data": {
			"repository": {
				"discussion": {
					"id": "D_kwDOA",
					"title": "Test Discussion",
					"url": "https://github.com/owner/repo/discussions/1",
					"state": "OPEN",
					"body": "This is a test",
					"createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
					"author": {"login": "testuser"},
					"category": {"name": "General"},
					"labels": {"nodes": []},
					"comments": {"nodes": []}
				}
			}
		}
	}`)

	disc, err := parseDiscussionResponse(body, 1)
	if err != nil {
		t.Fatalf("parseDiscussionResponse failed: %v", err)
	}

	if disc.GitResource.Title != "Test Discussion" {
		t.Errorf("expected title 'Test Discussion', got %q", disc.GitResource.Title)
	}
	if disc.GitResource.Author != "testuser" {
		t.Errorf("expected author 'testuser', got %q", disc.GitResource.Author)
	}
	if disc.GitResource.State != "open" {
		t.Errorf("expected state 'open', got %q", disc.GitResource.State)
	}
}

func TestParseDiscussionResponse_WithComments(t *testing.T) {
	body := []byte(`{
		"data": {
			"repository": {
				"discussion": {
					"id": "D_kwDOA",
					"title": "T",
					"url": "https://github.com/owner/repo/discussions/1",
					"state": "OPEN",
					"body": "body",
					"createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
					"author": {"login": "user1"},
					"category": {"name": "General"},
					"labels": {"nodes": []},
					"comments": {
						"nodes": [
							{
								"id": "DC_1",
								"body": "First comment",
								"createdAt": "2024-01-02T00:00:00Z",
								"updatedAt": "2024-01-02T00:00:00Z",
								"author": {"login": "user2"},
								"isAnswer": true,
								"reactions": {"nodes": [{"content": "THUMBS_UP"}, {"content": "HEART"}]},
								"replies": {
									"nodes": [
										{
											"id": "DC_2",
											"body": "Reply to first",
											"createdAt": "2024-01-03T00:00:00Z",
											"updatedAt": "2024-01-03T00:00:00Z",
											"author": {"login": "user3"},
											"reactions": {"nodes": []},
											"replies": {"nodes": []}
										}
									]
								}
							}
						]
					}
				}
			}
		}
	}`)

	disc, err := parseDiscussionResponse(body, 1)
	if err != nil {
		t.Fatalf("parseDiscussionResponse failed: %v", err)
	}

	if len(disc.Comments) != 2 {
		t.Fatalf("expected 2 comments (1 parent + 1 reply), got %d", len(disc.Comments))
	}

	// First comment should be the parent
	if disc.Comments[0].Author != "user2" {
		t.Errorf("expected first comment author 'user2', got %q", disc.Comments[0].Author)
	}
	if !disc.Comments[0].IsAnswer {
		t.Error("expected first comment to be answer")
	}
	if disc.Comments[0].Reactions.ThumbsUp != 1 {
		t.Errorf("expected 1 thumbs up, got %d", disc.Comments[0].Reactions.ThumbsUp)
	}
	if disc.Comments[0].Reactions.Heart != 1 {
		t.Errorf("expected 1 heart, got %d", disc.Comments[0].Reactions.Heart)
	}

	// Reply should have ParentID set
	if disc.Comments[1].ParentID == nil {
		t.Error("expected reply to have ParentID, got nil")
	}
}

func TestParseDiscussionResponse_EmptyData(t *testing.T) {
	// Test empty discussion (no data)
	_, err := parseDiscussionResponse([]byte(`{"data": {}}`), 1)
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"OPEN", "open"},
		{"Closed", "closed"},
		{"LOCKED", "locked"},
		{"", ""},
		{"already-lower", "already-lower"},
	}
	for _, tt := range tests {
		got := toLower(tt.input)
		if got != tt.expected {
			t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
