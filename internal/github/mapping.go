package github

import (
	"time"

	"github.com/google/go-github/v69/github"
)

// mapIssue 将 go-github 的 Issue 类型映射为内部模型。
func mapIssue(gi *github.Issue, comments []*github.IssueComment) *Issue {
	if gi == nil {
		return nil
	}

	issue := &Issue{
		GitResource: mapGitResource(gi.GetNumber(), gi.GetTitle(), gi.GetHTMLURL(),
			gi.GetUser().GetLogin(), gi.GetState(), gi.GetBody(),
			gi.GetCreatedAt().Time, gi.GetUpdatedAt().Time,
			mapLabels(gi.Labels), mapMilestone(gi.Milestone)),
		Comments: mapIssueComments(comments),
	}
	return issue
}

// mapPullRequest 将 go-github 的 PullRequest 类型映射为内部模型。
func mapPullRequest(pr *github.PullRequest, issueComments []*github.IssueComment, reviewComments []*github.PullRequestComment) *PullRequest {
	if pr == nil {
		return nil
	}

	result := &PullRequest{
		GitResource: mapGitResource(pr.GetNumber(), pr.GetTitle(), pr.GetHTMLURL(),
			pr.GetUser().GetLogin(), pr.GetState(), pr.GetBody(),
			pr.GetCreatedAt().Time, pr.GetUpdatedAt().Time,
			mapLabels(pr.Labels), mapMilestone(pr.Milestone)),
		Comments: make([]Comment, 0, len(issueComments)+len(reviewComments)),
	}

	// 合并 Issue 评论和 Review 评论
	for _, c := range issueComments {
		result.Comments = append(result.Comments, mapIssueComment(c))
	}
	for _, c := range reviewComments {
		result.Comments = append(result.Comments, mapPullReviewComment(c))
	}

	// 按创建时间正序排序
	sortCommentsByCreatedAt(result.Comments)

	return result
}

// mapGitResource 映射共享字段。
func mapGitResource(number int, title, htmlURL, author, state, body string, createdAt, updatedAt time.Time, labels []string, milestone string) GitHubResource {
	return GitHubResource{
		Number:    number,
		Title:     title,
		URL:       htmlURL,
		Author:    author,
		State:     state,
		Body:      body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Labels:    labels,
		Milestone: milestone,
	}
}

// mapLabels 从 go-github Label 切片提取名称。
func mapLabels(labels []*github.Label) []string {
	if len(labels) == 0 {
		return nil
	}
	result := make([]string, 0, len(labels))
	for _, l := range labels {
		if l != nil {
			result = append(result, l.GetName())
		}
	}
	return result
}

// mapMilestone 从 go-github Milestone 提取名称。
func mapMilestone(m *github.Milestone) string {
	if m == nil {
		return ""
	}
	return m.GetTitle()
}

// mapIssueComment 映射 Issue/PR 常规评论。
func mapIssueComment(c *github.IssueComment) Comment {
	if c == nil {
		return Comment{}
	}

	comment := Comment{
		ID:        c.GetID(),
		Author:    c.GetUser().GetLogin(),
		Body:      c.GetBody(),
		CreatedAt: c.GetCreatedAt().Time,
		UpdatedAt: c.GetUpdatedAt().Time,
		Reactions: mapReactions(c.Reactions),
	}
	return comment
}

// mapIssueComments 映射 Issue 评论列表。
func mapIssueComments(comments []*github.IssueComment) []Comment {
	if len(comments) == 0 {
		return nil
	}
	result := make([]Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, mapIssueComment(c))
	}
	return result
}

// mapPullReviewComment 映射 PR Review 评论。
func mapPullReviewComment(c *github.PullRequestComment) Comment {
	if c == nil {
		return Comment{}
	}
	return Comment{
		ID:        c.GetID(),
		Author:    c.GetUser().GetLogin(),
		Body:      c.GetBody(),
		CreatedAt: c.GetCreatedAt().Time,
		UpdatedAt: c.GetUpdatedAt().Time,
		FilePath:  c.GetPath(),
		Line:      c.GetLine(),
	}
}

// mapReactions 映射 Reactions 统计。
func mapReactions(r *github.Reactions) ReactionCount {
	if r == nil {
		return ReactionCount{}
	}
	return ReactionCount{
		ThumbsUp:   r.GetPlusOne(),
		ThumbsDown: r.GetMinusOne(),
		Laugh:      r.GetLaugh(),
		Hooray:     r.GetHooray(),
		Confused:   r.GetConfused(),
		Heart:      r.GetHeart(),
		Rocket:     r.GetRocket(),
		Eyes:       r.GetEyes(),
	}
}

// sortCommentsByCreatedAt 按创建时间正序排序评论。
func sortCommentsByCreatedAt(comments []Comment) {
	for i := 0; i < len(comments); i++ {
		for j := i + 1; j < len(comments); j++ {
			if comments[j].CreatedAt.Before(comments[i].CreatedAt) {
				comments[i], comments[j] = comments[j], comments[i]
			}
		}
	}
}
