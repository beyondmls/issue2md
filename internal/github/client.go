package github

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/google/go-github/v69/github"
)

// Client 封装 GitHub REST v3 和 GraphQL v4 API。
//
// REST 请求通过 go-github 库完成，GraphQL 请求（Discussions）
// 通过原始 HTTP POST 完成。
type Client struct {
	restClient *github.Client
	token      string
}

// NewClient 创建一个 GitHub API 客户端。
// token 为空字符串时以未认证模式运行（速率限制较低）。
func NewClient(token string) *Client {
	var rest *github.Client
	if token != "" {
		rest = github.NewClient(nil).WithAuthToken(token)
	} else {
		rest = github.NewClient(nil)
	}
	return &Client{
		restClient: rest,
		token:      token,
	}
}

// FetchIssue 获取 Issue 详情及所有评论。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	gi, _, err := c.restClient.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetching issue %s/%s#%d: %w", owner, repo, number, err)
	}

	comments, err := c.listAllIssueComments(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetching issue comments %s/%s#%d: %w", owner, repo, number, err)
	}

	return mapIssue(gi, comments), nil
}

// FetchPullRequest 获取 PR 详情及所有评论。
//
// 并发获取 Review Comments 和常规 Issue Comments，
// 合并后按创建时间正序排序。
func (c *Client) FetchPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	pr, _, err := c.restClient.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("fetching PR %s/%s#%d: %w", owner, repo, number, err)
	}

	var (
		issueComments  []*github.IssueComment
		reviewComments []*github.PullRequestComment
	)

	var wg sync.WaitGroup
	var icErr, rcErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		issueComments, icErr = c.listAllIssueComments(ctx, owner, repo, number)
	}()
	go func() {
		defer wg.Done()
		reviewComments, rcErr = c.listAllPullReviewComments(ctx, owner, repo, number)
	}()
	wg.Wait()

	if icErr != nil {
		return nil, fmt.Errorf("fetching PR issue comments %s/%s#%d: %w", owner, repo, number, icErr)
	}
	if rcErr != nil {
		return nil, fmt.Errorf("fetching PR review comments %s/%s#%d: %w", owner, repo, number, rcErr)
	}

	result := mapPullRequest(pr, issueComments, reviewComments)
	sort.Slice(result.Comments, func(i, j int) bool {
		return result.Comments[i].CreatedAt.Before(result.Comments[j].CreatedAt)
	})
	return result, nil
}

// FetchDiscussion 获取 Discussion 详情及所有评论（含层级结构）。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (*Discussion, error) {
	return fetchDiscussionGraphQL(ctx, c.token, owner, repo, number)
}

// listAllIssueComments 遍历所有页面获取 Issue/PR 的 Issue 评论。
func (c *Client) listAllIssueComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
	var all []*github.IssueComment
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := c.restClient.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

// listAllPullReviewComments 遍历所有页面获取 PR 的 Review 评论。
func (c *Client) listAllPullReviewComments(ctx context.Context, owner, repo string, number int) ([]*github.PullRequestComment, error) {
	var all []*github.PullRequestComment
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := c.restClient.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}
