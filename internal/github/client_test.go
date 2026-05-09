package github

import (
	"context"
	"os"
	"testing"
)

func skipIfNoToken(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}
	return token
}

func TestFetchIssue_Real(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	issue, err := client.FetchIssue(context.Background(), "golang", "go", 1)
	if err != nil {
		t.Fatalf("FetchIssue failed: %v", err)
	}

	if issue.GitResource.Number != 1 {
		t.Errorf("expected number 1, got %d", issue.GitResource.Number)
	}
	if issue.GitResource.Author == "" {
		t.Error("expected non-empty author")
	}
	if issue.GitResource.Title == "" {
		t.Error("expected non-empty title")
	}
	if issue.GitResource.State != "open" && issue.GitResource.State != "closed" {
		t.Errorf("expected state open or closed, got %q", issue.GitResource.State)
	}
}

func TestFetchPullRequest_Real(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	pr, err := client.FetchPullRequest(context.Background(), "golang", "go", 1)
	if err != nil {
		t.Fatalf("FetchPullRequest failed: %v", err)
	}

	if pr.GitResource.Number != 1 {
		t.Errorf("expected number 1, got %d", pr.GitResource.Number)
	}
	if pr.GitResource.Author == "" {
		t.Error("expected non-empty author")
	}

	// PRs should have at least the PR body comments available
	t.Logf("PR #1 has %d comments", len(pr.Comments))
}

func TestFetchPullRequest_MergesComments(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	// Use a known PR with both issue comments and review comments
	pr, err := client.FetchPullRequest(context.Background(), "golang", "go", 10)
	if err != nil {
		t.Fatalf("FetchPullRequest failed: %v", err)
	}

	// Verify comments are sorted by CreatedAt
	for i := 1; i < len(pr.Comments); i++ {
		if pr.Comments[i].CreatedAt.Before(pr.Comments[i-1].CreatedAt) {
			t.Errorf("comments not sorted by CreatedAt: index %d (%v) before %d (%v)",
				i-1, pr.Comments[i-1].CreatedAt, i, pr.Comments[i].CreatedAt)
		}
	}

	t.Logf("PR #10 has %d comments total", len(pr.Comments))
}

func TestFetchIssue_NotFound(t *testing.T) {
	token := skipIfNoToken(t)

	client := NewClient(token)
	_, err := client.FetchIssue(context.Background(), "golang", "go", 99999999)
	if err == nil {
		t.Fatal("expected error for non-existent issue, got nil")
	}
	t.Logf("Got expected error: %v", err)
}

func TestNewClient_NoToken(t *testing.T) {
	client := NewClient("")
	if client == nil {
		t.Fatal("NewClient with empty token should not return nil")
	}
}

func TestNewClient_WithToken(t *testing.T) {
	client := NewClient("test-token")
	if client == nil {
		t.Fatal("NewClient with token should not return nil")
	}
}
