# API Sketch — 核心包接口设计

> 本文档定义了 `internal/github` 和 `internal/converter` 两个核心包的对外接口签名，作为编码阶段的参考。
> 遵循宪法原则：无全局变量，依赖显式注入，标准库优先。

---

## 1. internal/parser — URL 解析

职责：将 GitHub URL 解析为结构化信息。独立于其他包。

```go
package parser

// ResourceType 表示 GitHub 资源类型
type ResourceType string

const (
    ResourceIssue      ResourceType = "issue"
    ResourcePull       ResourceType = "pull"
    ResourceDiscussion ResourceType = "discussion"
)

// URLInfo 是 URL 解析结果
type URLInfo struct {
    Owner  string
    Repo   string
    Type   ResourceType
    Number int
}

// ParseURL 解析 GitHub URL，自动识别资源类型。
// 支持格式:
//   https://github.com/{owner}/{repo}/issues/{number}
//   https://github.com/{owner}/{repo}/pull/{number}
//   https://github.com/{owner}/{repo}/discussions/{number}
// 尾部斜杠和无关查询参数不影响解析。
// 返回 ErrInvalidURL（可导出）当 URL 不匹配任何已知模式时。
func ParseURL(rawURL string) (*URLInfo, error)
```

---

## 2. internal/github — GitHub API 客户端

职责：通过 REST v3 和 GraphQL v4 获取 GitHub 数据。请求 `internal/parser` 解析的 URL 来获取数据。

### 2.1 数据模型 (models.go)

```go
package github

import "time"

// Comment 表示一条评论（Issue 评论、PR Review 评论、Discussion 回复）
type Comment struct {
    ID        int64
    Author    string       // GitHub 用户名
    Body      string       // Markdown 正文
    CreatedAt time.Time
    UpdatedAt time.Time
    Reactions ReactionCount // 当 API 返回时填充

    // 以下字段仅在特定资源类型中有值

    // Discussion Answer 标记
    IsAnswer  bool
    // Discussion 父评论 ID（用于层级结构）
    ParentID  *int64
    // PR Review 评论的关联文件
    FilePath  string
    // PR Review 评论的关联行号
    Line      int
}

// ReactionCount 存储 Reactions 统计
type ReactionCount struct {
    ThumbsUp   int
    ThumbsDown int
    Laugh      int
    Hooray     int
    Confused   int
    Heart      int
    Rocket     int
    Eyes       int
}

// Issue 表示一个 GitHub Issue
type Issue struct {
    Number    int
    Title     string
    URL       string
    Author    string
    State     string       // "open" 或 "closed"
    Body      string       // Markdown 正文
    CreatedAt time.Time
    UpdatedAt time.Time
    Labels    []string
    Milestone string
    Comments  []Comment
}

// PullRequest 表示一个 GitHub Pull Request（不含 Diff 信息）
type PullRequest struct {
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
    Comments  []Comment  // 合并了 Issue-style 和 Review 评论
}

// Discussion 表示一个 GitHub Discussion
type Discussion struct {
    Number    int
    Title     string
    URL       string
    Author    string
    State     string       // "open" 或 "closed"
    Body      string
    CreatedAt time.Time
    UpdatedAt time.Time
    Labels    []string
    Milestone string
    Comments  []Comment    // 保留 ParentID 层级关系
}
```

### 2.2 客户端接口 (client.go)

```go
package github

// NewClient 创建一个 GitHub API 客户端。
// token 为空字符串时以未认证模式运行（rate limit 较低）。
func NewClient(token string) *Client

// Client 封装 GitHub REST v3 和 GraphQL v4 API。
type Client struct {
    token   string
    baseURL string       // 便于测试时 mock（默认 "https://api.github.com"）
    http    *http.Client
}

// FetchIssue 获取 Issue 详情及所有评论。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (*Issue, error)

// FetchPullRequest 获取 PR 详情及所有评论（含 Review 评论，不含 Diff）。
func (c *Client) FetchPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error)

// FetchDiscussion 获取 Discussion 详情及所有评论（含层级结构）。
// 使用 GraphQL v4 API。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (*Discussion, error)
```

### 2.3 内部辅助方法（不暴露）

```go
// --- REST 辅助 ---
func (c *Client) getREST(ctx context.Context, path string) ([]byte, error)
func (c *Client) fetchIssueComments(ctx context.Context, owner, repo string, number int) ([]Comment, error)
func (c *Client) fetchPRReviewComments(ctx context.Context, owner, repo string, number int) ([]Comment, error)

// --- GraphQL 辅助 ---
func (c *Client) postGraphQL(ctx context.Context, query string, vars map[string]any) ([]byte, error)
func (c *Client) fetchDiscussionGraphQL(ctx context.Context, owner, repo string, number int) (*Discussion, error)
```

---

## 3. internal/converter — Markdown 渲染

职责：将 `internal/github` 的数据模型渲染为 Markdown 字符串（含 YAML Frontmatter）。不依赖网络。

```go
package converter

import "github.com/bigwhite/issue2md/internal/github"

// Option 是渲染选项
type Option struct {
    EnableReactions bool     // 是否渲染 Reactions 统计
    EnableUserLinks bool     // 是否将 @user 渲染为 GitHub 链接
}

// New 创建一个转换器。
// 所有配置通过 Option 注入，不依赖全局状态。
func New(opts Option) *Converter

// Converter 负责将 GitHub 数据模型转换为 Markdown。
type Converter struct {
    opts Option
}

// ConvertIssue 将 Issue 渲染为 Markdown（含 YAML Frontmatter）。
func (c *Converter) ConvertIssue(issue *github.Issue) (string, error)

// ConvertPullRequest 将 PR 渲染为 Markdown（含 YAML Frontmatter）。
func (c *Converter) ConvertPullRequest(pr *github.PullRequest) (string, error)

// ConvertDiscussion 将 Discussion 渲染为 Markdown（含 YAML Frontmatter）。
func (c *Converter) ConvertDiscussion(disc *github.Discussion) (string, error)
```

### 3.1 内部辅助函数（不暴露）

```go
// renderFrontmatter 生成 YAML Frontmatter
func renderFrontmatter(title, url, author, state, createdAt, updatedAt string, number int, rtype string, labels []string, milestone string) string

// renderComments 渲染评论列表
func renderComments(comments []github.Comment, opts Option) string

// renderComment 渲染单条评论
func renderComment(comment github.Comment, opts Option) string

// renderReactions 渲染 Reactions 行
func renderReactions(r github.ReactionCount) string

// renderUserLinks 将 @mentions 替换为 Markdown 链接
func renderUserLinks(body string) string
```

---

## 4. internal/config — 配置管理

```go
package config

// GetGitHubToken 从环境变量 GITHUB_TOKEN 获取 Token。
// 环境变量不存在时返回空字符串。
func GetGitHubToken() string
```

---

## 5. internal/cli — 命令行参数

```go
package cli

// Flags 存储命令行解析结果
type Flags struct {
    EnableReactions bool
    EnableUserLinks bool
}

// Parse 解析 os.Args[1:]。
// 返回 Flags、url、outputFile（可能为空字符串）、错误。
func Parse(args []string) (*Flags, url string, outputFile string, err error)
```

---

## 6. 包间依赖关系

```
cmd/issue2md/main.go
    │ 依赖
    ├── internal/cli        (解析参数)
    ├── internal/config     (读取 Token)
    ├── internal/parser     (解析 URL)
    ├── internal/github     (获取数据)
    └── internal/converter  (渲染 Markdown)

internal/github
    └── 依赖 internal/parser (可选: URLInfo 作为参数结构复用)

internal/converter
    └── 依赖 internal/github (仅模型类型)

internal/cli
    └── 无内部依赖

internal/parser
    └── 无内部依赖

internal/config
    └── 无内部依赖
```

---

## 7. main.go 编排逻辑（伪代码）

```go
func main() {
    // 1. 解析命令行
    flags, rawURL, outputFile, err := cli.Parse(os.Args[1:])

    // 2. 读取 Token
    token := config.GetGitHubToken()

    // 3. 解析 URL
    info, err := parser.ParseURL(rawURL)

    // 4. 创建 API 客户端并获取数据
    client := github.NewClient(token)
    ctx := context.Background()

    var data any
    switch info.Type {
    case parser.ResourceIssue:
        data, err = client.FetchIssue(ctx, info.Owner, info.Repo, info.Number)
    case parser.ResourcePull:
        data, err = client.FetchPullRequest(ctx, info.Owner, info.Repo, info.Number)
    case parser.ResourceDiscussion:
        data, err = client.FetchDiscussion(ctx, info.Owner, info.Repo, info.Number)
    }

    // 5. 转换为 Markdown
    conv := converter.New(converter.Option{...})
    var md string
    switch v := data.(type) {
    case *github.Issue:
        md, err = conv.ConvertIssue(v)
    case *github.PullRequest:
        md, err = conv.ConvertPullRequest(v)
    case *github.Discussion:
        md, err = conv.ConvertDiscussion(v)
    }

    // 6. 输出
    if outputFile != "" {
        os.WriteFile(outputFile, []byte(md), 0644)
    } else {
        os.Stdout.WriteString(md)
    }
}
```
