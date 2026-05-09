package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigwhite/issue2md/internal/github"
)

// renderFrontmatter 生成 YAML Frontmatter 字符串。
func renderFrontmatter(r github.GitHubResource, rtype string) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", r.Title)
	fmt.Fprintf(&b, "url: %s\n", r.URL)
	fmt.Fprintf(&b, "number: %d\n", r.Number)
	fmt.Fprintf(&b, "type: %s\n", rtype)
	fmt.Fprintf(&b, "author: %s\n", r.Author)
	fmt.Fprintf(&b, "state: %s\n", r.State)
	fmt.Fprintf(&b, "created_at: %q\n", r.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "updated_at: %q\n", r.UpdatedAt.Format(time.RFC3339))

	if len(r.Labels) == 0 {
		b.WriteString("labels: []\n")
	} else {
		b.WriteString("labels:\n")
		for _, l := range r.Labels {
			fmt.Fprintf(&b, "  - %s\n", l)
		}
	}

	if r.Milestone == "" {
		b.WriteString("milestone: \"\"\n")
	} else {
		fmt.Fprintf(&b, "milestone: %s\n", r.Milestone)
	}

	b.WriteString("---")
	return b.String()
}

// renderTitle 生成 Markdown 标题行。
func renderTitle(title string, number int) string {
	return fmt.Sprintf("# %s (#%d)", title, number)
}
