package github

import "time"

// ReactionCount 对应 GitHub API reactions 的统计值。
type ReactionCount struct {
	ThumbsUp   int `json:"+1"`
	ThumbsDown int `json:"-1"`
	Laugh      int `json:"laugh"`
	Hooray     int `json:"hooray"`
	Confused   int `json:"confused"`
	Heart      int `json:"heart"`
	Rocket     int `json:"rocket"`
	Eyes       int `json:"eyes"`
}

// Empty 返回 true 当没有任何 reaction 时。
func (r ReactionCount) Empty() bool {
	return r.ThumbsUp == 0 && r.ThumbsDown == 0 && r.Laugh == 0 &&
		r.Hooray == 0 && r.Confused == 0 && r.Heart == 0 &&
		r.Rocket == 0 && r.Eyes == 0
}

// Comment 表示 GitHub 上的一条评论。
//
// 适用于三种资源:
//   - Issue: 常规 Issue 评论
//   - Pull Request: 包含 Review 评论 (FilePath/Line 有值) 和常规评论
//   - Discussion: 包含层级关系 (ParentID) 和 Answer 标记
type Comment struct {
	ID        int64
	Author    string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Reactions ReactionCount

	// Discussion: 该评论是否为被采纳的 Answer
	IsAnswer bool
	// Discussion: 父评论 ID（为 nil 表示顶层评论）
	ParentID *int64

	// PR Review Comment: 关联的文件路径
	FilePath string
	// PR Review Comment: 关联的行号
	Line int
}

// GitHubResource 是 Issue / PR / Discussion 共享的字段。
type GitHubResource struct {
	Number    int
	Title     string
	URL       string
	Author    string
	State     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Labels    []string
	Milestone string
}

// Issue 表示一个 GitHub Issue 及其评论。
type Issue struct {
	GitResource GitHubResource
	Comments    []Comment
}

// PullRequest 表示一个 GitHub Pull Request。
// 不包含 Diff 信息，只包含描述和所有评论。
type PullRequest struct {
	GitResource GitHubResource
	Comments    []Comment
}

// Discussion 表示一个 GitHub Discussion 及其评论层级。
type Discussion struct {
	GitResource GitHubResource
	Comments    []Comment
}
