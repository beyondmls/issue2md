package converter

import (
	"errors"

	"github.com/bigwhite/issue2md/internal/github"
)

// Option 控制渲染行为。
type Option struct {
	EnableReactions bool
	EnableUserLinks bool
}

// Converter 负责将 GitHub 数据模型渲染为 Markdown（含 YAML Frontmatter）。
type Converter struct {
	opts Option
}

// New 创建一个转换器。
func New(opts Option) *Converter {
	return &Converter{opts: opts}
}

// ConvertIssue 将 Issue 渲染为 Markdown。
func (c *Converter) ConvertIssue(issue *github.Issue) (string, error) {
	if issue == nil {
		return "", errors.New("nil issue")
	}
	return c.render(issue.GitResource, "issue", issue.Comments), nil
}

// ConvertPullRequest 将 Pull Request 渲染为 Markdown。
func (c *Converter) ConvertPullRequest(pr *github.PullRequest) (string, error) {
	if pr == nil {
		return "", errors.New("nil pull request")
	}
	return c.render(pr.GitResource, "pull", pr.Comments), nil
}

// ConvertDiscussion 将 Discussion 渲染为 Markdown。
func (c *Converter) ConvertDiscussion(disc *github.Discussion) (string, error) {
	if disc == nil {
		return "", errors.New("nil discussion")
	}
	return c.render(disc.GitResource, "discussion", disc.Comments), nil
}

// render 是通用的渲染入口。
func (c *Converter) render(r github.GitHubResource, rtype string, comments []github.Comment) string {
	// 对描述正文中的 @mentions 做链接转换
	body := r.Body
	if c.opts.EnableUserLinks {
		body = renderUserLinks(body)
	}

	// 构建 comments 区域
	commentOpts := c.opts
	commentsStr := renderComments(comments, commentOpts)

	// 构建完整 body（不含 frontmatter）
	fullBody := renderFullBody(github.GitHubResource{
		Title:  r.Title,
		Number: r.Number,
		Body:   body,
	}, commentsStr)

	// 生成 frontmatter
	fm := renderFrontmatter(r, rtype)

	return fm + "\n\n" + fullBody
}

// renderFullBody 渲染主体内容（不含 frontmatter）。
func renderFullBody(r github.GitHubResource, comments string) string {
	title := renderTitle(r.Title, r.Number)
	if comments == "" {
		return title + "\n\n" + r.Body
	}
	return title + "\n\n" + r.Body + comments
}
