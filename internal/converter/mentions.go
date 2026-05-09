package converter

import "regexp"

// mentionRe 匹配 @username 模式的 GitHub 用户提及。
// GitHub 用户名规则: 字母、数字、连字符，不能以连字符开头或结尾。
var mentionRe = regexp.MustCompile(`@([a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?)`)

// renderUserLinks 将 @username 替换为指向其 GitHub 主页的 Markdown 链接。
func renderUserLinks(body string) string {
	return mentionRe.ReplaceAllString(body, "[@$1](https://github.com/$1)")
}
