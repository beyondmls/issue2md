package converter

import (
	"fmt"
	"strings"

	"github.com/bigwhite/issue2md/internal/github"
)

// renderComments 渲染评论列表。
// 如果没有评论，返回空字符串。
func renderComments(comments []github.Comment, opts Option) string {
	if len(comments) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n---\n\n## Comments\n")

	for i, c := range comments {
		b.WriteString("\n")
		b.WriteString(renderCommentHeader(c))
		b.WriteString("\n\n")

		body := c.Body
		if opts.EnableUserLinks {
			body = renderUserLinks(body)
		}

		if c.IsAnswer {
			b.WriteString(blockquote(body))
		} else {
			b.WriteString(body)
		}

		if opts.EnableReactions && !c.Reactions.Empty() {
			b.WriteString("\n\n")
			b.WriteString(renderReactions(c.Reactions))
		}

		if i < len(comments)-1 {
			b.WriteString("\n\n---")
		}
	}

	return b.String()
}

// renderCommentHeader 渲染单条评论的标题行。
func renderCommentHeader(c github.Comment) string {
	prefix := "[Comment]"
	if c.IsAnswer {
		prefix = "✅ Answer"
	} else if c.FilePath != "" {
		prefix = "[Review Comment]"
	}

	timeStr := c.CreatedAt.Format("2006-01-02T15:04:05Z")

	var b strings.Builder
	if prefix == "[Review Comment]" {
		fmt.Fprintf(&b, "### %s by **%s** on `%s`", prefix, c.Author, c.FilePath)
		if c.Line > 0 {
			fmt.Fprintf(&b, " (line %d)", c.Line)
		}
		fmt.Fprintf(&b, " — %s", timeStr)
	} else {
		fmt.Fprintf(&b, "### %s by **%s** — %s", prefix, c.Author, timeStr)
	}

	return b.String()
}

// renderReactions 渲染 Reactions 统计行。
func renderReactions(r github.ReactionCount) string {
	var parts []string
	addIf := func(count int, emoji string) {
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", emoji, count))
		}
	}

	addIf(r.ThumbsUp, "👍")
	addIf(r.ThumbsDown, "👎")
	addIf(r.Laugh, "😄")
	addIf(r.Hooray, "🎉")
	addIf(r.Confused, "😕")
	addIf(r.Heart, "❤️")
	addIf(r.Rocket, "🚀")
	addIf(r.Eyes, "👀")

	return "> Reactions: " + strings.Join(parts, " ")
}

// blockquote 将文本包装为 Markdown 块引用。
func blockquote(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return strings.Join(lines, "\n")
}
