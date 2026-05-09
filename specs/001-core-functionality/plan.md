# issue2md — 技术实现方案 (Plan)

> **版本**: v0.1 (MVP)
> **对应 Spec**: `specs/001-core-functionality/spec.md`
> **最后更新**: 2026-05-09

---

## 1. 技术上下文总结

### 1.1 技术选型

| 层 | 技术 | 理由 |
|---|---|---|
| 语言 | Go 1.26.3 (`go.mod` 定义) | 项目已有定义 |
| REST API 客户端 | `google/go-github` v69+ | GitHub 官方推荐的 Go 客户端，内置分页、速率限制、鉴权处理 |
| GraphQL API | `net/http` + 手动构造 JSON body | go-github 对 Discussion 无类型安全的支持；手动构造 Query 更灵活 |
| CLI 框架 | `flag` 标准库 | 宪法 1.2: 标准库优先 |
| JSON 处理 | `encoding/json` 标准库 | 宪法 1.2 |
| Markdown 渲染 | `strings.Builder` + `fmt` 标准库 | **零外部依赖**渲染，手动字符串拼接 |
| 测试 | `testing` 标准库 | 表格驱动测试 |
| Web (预留) | `net/http` 标准库 | 宪法 1.2 |

### 1.2 依赖策略

```
google/go-github v69+
  └── golang.org/x/oauth2    (间接依赖，用于 Token 鉴权)
```

这是 MVP 阶段**唯一**的第三方依赖。理由：
- go-github 处理翻页、HTTP 错误码映射、速率限制等数百行样板代码
- 宪法 1.2 允许在标准库无法高效满足的场景下引入成熟的外部库
- `internal/converter` 和 `internal/parser` 不依赖 go-github，保持了零外部依赖的纯净

### 1.3 api-sketch 调整说明

引入 go-github 后，`internal/github` 包的设计调整为：

```
internal/github/
  ├── client.go          # NewClient, FetchIssue, FetchPullRequest, FetchDiscussion
  ├── client_test.go     # 集成测试
  ├── models.go          # 内部模型 (Issue, PullRequest, Discussion, Comment)
  └── mapping.go         # go-github 类型 → 内部模型的映射逻辑
```

`client.go` 使用 go-github 完成 REST 调用，但**对外只暴露我们自己的模型类型**。

---

## 2. "合宪性"审查

### 第一条：简单性原则

| 条款 | 方案做法 | 状态 |
|---|---|---|
| 1.1 YAGNI | 只实现 spec.md 中定义的功能，不做批量转换、Web UI 等 | ✅ |
| 1.2 标准库优先 | `flag` / `net/http` / `encoding/json` / `strings` — 只引入 go-github 这一个外部依赖 | ✅ |
| 1.3 反过度工程 | 不定义 interface 层(除非测试需要 mock)，直接用 struct + 方法；不抽象"通用的 HTTP 客户端" | ✅ |

### 第二条：测试先行铁律

| 条款 | 方案做法 | 状态 |
|---|---|---|
| 2.1 TDD 循环 | 每个实现阶段：先写测试 → 运行失败 → 实现 → 验证通过 | ✅ |
| 2.2 表格驱动 | URL 解析、Markdown 渲染、CLI 解析均使用表格驱动测试 | ✅ |
| 2.3 拒绝 Mocks | `internal/github` 的集成测试使用真实的 `github.com` API；单元测试层使用纯函数无网络依赖 | ✅ |

### 第三条：明确性原则

| 条款 | 方案做法 | 状态 |
|---|---|---|
| 3.1 错误处理 | 所有错误使用 `fmt.Errorf("...: %w", err)` 包装；错误输出到 stderr；exit code 0/1 | ✅ |
| 3.2 无全局变量 | 所有依赖(Client, Converter, 配置)通过函数参数和结构体显式注入 | ✅ |

---

## 3. 项目结构 (详细版)

```
.
├── cmd/
│   └── issue2md/
│       └── main.go                # CLI 入口：编排 cli → parser → github → converter → output
│
├── internal/
│   ├── cli/
│   │   ├── parse.go               # Parse(): flags + url + output_file 解析
│   │   └── parse_test.go          # 表格驱动测试
│   │
│   ├── config/
│   │   └── token.go               # GetGitHubToken(): 读取 GITHUB_TOKEN 环境变量
│   │
│   ├── parser/
│   │   ├── url.go                 # ParseURL(): GitHub URL → URLInfo
│   │   ├── url_test.go            # 表格驱动测试 (合法/非法/边界 URL)
│   │   └── errors.go              # ErrInvalidURL 定义
│   │
│   ├── github/
│   │   ├── client.go              # NewClient(), FetchIssue/PR/Discussion
│   │   ├── client_test.go         # 集成测试 (真实 API)
│   │   ├── models.go              # Issue, PullRequest, Discussion, Comment, ReactionCount
│   │   ├── mapping.go             # go-github types → 内部 models
│   │   └── graphql.go             # Discussion 的 GraphQL query + 响应解析
│   │
│   ├── converter/
│   │   ├── converter.go           # New(), ConvertIssue/PR/Discussion
│   │   ├── converter_test.go      # 表格驱动测试 (纯函数，无网络)
│   │   ├── frontmatter.go         # renderFrontmatter: YAML 头部生成
│   │   ├── comments.go            # renderComments/renderComment/renderReactions
│   │   └── mentions.go            # renderUserLinks: @user → [user](url)
│   │
│   └── web/                       # 预留：未来 Web UI 的 handler
│       └── handler.go
│
├── web/
│   ├── templates/                 # 预留：Web 模板
│   └── static/                    # 预留：静态资源
│
├── specs/
│   └── 001-core-functionality/
│       ├── spec.md                # 功能规格
│       ├── api-sketch.md          # 接口设计草图
│       └── plan.md                # 本文件
│
├── Makefile                       # 构建与测试命令
├── go.mod
├── CLAUDE.md
└── constitution.md
```

### 包依赖图

```
cmd/issue2md/main.go
  ├── internal/cli          (无内部依赖)
  ├── internal/config       (无内部依赖)
  ├── internal/parser       (无内部依赖)
  ├── internal/github       (外部: google/go-github)
  └── internal/converter    (依赖 internal/github 的模型)
```

---

## 4. 核心数据结构

以下结构定义在 `internal/github/models.go` 中，是包间契约的中心。

```go
package github

import "time"

// ResourceType 区分三种 GitHub 资源
type ResourceType string

const (
    ResourceIssue      ResourceType = "issue"
    ResourcePull       ResourceType = "pull"
    ResourceDiscussion ResourceType = "discussion"
)

// ReactionCount 对应 GitHub API reactions 的 +1, -1, laugh, hooray, confused, heart, rocket, eyes
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

// Empty 检查是否有任何 reaction
func (r ReactionCount) Empty() bool { /* ... */ }

// Comment 表示一条评论
type Comment struct {
    ID        int64
    Author    string
    Body      string
    CreatedAt time.Time
    UpdatedAt time.Time
    Reactions ReactionCount

    // Discussion 相关
    IsAnswer  bool
    ParentID  *int64

    // PR Review Comment 相关
    FilePath string
    Line     int
}

// GitHubResource 是 Issue / PR / Discussion 的共同字段
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

// Issue 表示 GitHub Issue
type Issue struct {
    GitHubResource
    Comments []Comment
}

// PullRequest 表示 GitHub Pull Request (不含 diff)
type PullRequest struct {
    GitHubResource
    Comments []Comment
}

// Discussion 表示 GitHub Discussion
type Discussion struct {
    GitHubResource
    Comments []Comment
}
```

---

## 5. 接口设计

### 5.1 internal/cli

```go
package cli

// Params 是命令行解析结果
type Params struct {
    URL             string
    OutputFile      string   // 空字符串 = stdout
    EnableReactions bool
    EnableUserLinks bool
}

// Parse 解析 os.Args[1:], 返回 Params 或错误。
func Parse(args []string) (*Params, error)
```

### 5.2 internal/config

```go
package config

// GetGitHubToken 从 GITHUB_TOKEN 环境变量读取 token。
// 未设置时返回空字符串。
func GetGitHubToken() string
```

### 5.3 internal/parser

```go
package parser

import "errors"

var ErrInvalidURL = errors.New("invalid GitHub URL")

type URLInfo struct {
    Owner  string
    Repo   string
    Type   ResourceType
    Number int
}

type ResourceType string

const (
    ResourceIssue      ResourceType = "issue"
    ResourcePull       ResourceType = "pull"
    ResourceDiscussion ResourceType = "discussion"
)

func ParseURL(rawURL string) (*URLInfo, error)
```

### 5.4 internal/github

```go
package github

import "context"

type Client struct {
    // unexported: token, httpClient
}

func NewClient(token string, opts ...ClientOption) *Client

func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (*Issue, error)
func (c *Client) FetchPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error)
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (*Discussion, error)
```

**关于 `FetchPullRequest` 的内部实现**：

```go
// FetchPullRequest 并发获取 PR 详情 + Review 评论 + 常规 Issue 评论。
// 使用 errgroup 或 sync.WaitGroup 并发执行 3 个请求。
// 合并所有评论后按 CreatedAt 正序排序。
func (c *Client) FetchPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
    // 1. 获取 PR 详情 (go-github: PullRequests.Get)
    // 2. 并发获取:
    //    a. Review 评论 (go-github: PullRequests.ListComments)
    //    b. 常规 Issue 评论 (go-github: Issues.ListComments)
    // 3. 映射为内部模型
    // 4. 合并评论 + 按时间排序
    // 5. 返回 PullRequest
}
```

**关于 `FetchDiscussion` 的内部实现**：

```go
// FetchDiscussion 使用 GraphQL API 获取 Discussion 详情。
// go-github 不支持 Discussion API，使用 raw HTTP + GraphQL。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (*Discussion, error) {
    // 1. 构造 GraphQL query 请求 Discussion 及其 comments
    // 2. 发送 POST 到 https://api.github.com/graphql
    // 3. 解析响应为内部模型
    // 4. 保留评论的 ParentID 层级关系
    // 5. 标记 Answer 评论
    // 6. 返回 Discussion
}
```

### 5.5 internal/converter

```go
package converter

import "github.com/bigwhite/issue2md/internal/github"

type Option struct {
    EnableReactions bool
    EnableUserLinks bool
}

type Converter struct {
    opts Option
}

func New(opts Option) *Converter

func (c *Converter) ConvertIssue(issue *github.Issue) (string, error)
func (c *Converter) ConvertPullRequest(pr *github.PullRequest) (string, error)
func (c *Converter) ConvertDiscussion(disc *github.Discussion) (string, error)
```

### 5.6 cmd/issue2md/main.go 编排逻辑

```go
func main() io.Writer {
    params, err := cli.Parse(os.Args[1:])
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    info, err := parser.ParseURL(params.URL)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    token := config.GetGitHubToken()
    client := github.NewClient(token)
    ctx := context.Background()

    conv := converter.New(converter.Option{
        EnableReactions: params.EnableReactions,
        EnableUserLinks: params.EnableUserLinks,
    })

    var md string
    switch info.Type {
    case parser.ResourceIssue:
        issue, err := client.FetchIssue(ctx, info.Owner, info.Repo, info.Number)
        if err != nil { /* stderr + exit 1 */ }
        md, err = conv.ConvertIssue(issue)
    case parser.ResourcePull:
        pr, err := client.FetchPullRequest(ctx, info.Owner, info.Repo, info.Number)
        if err != nil { /* stderr + exit 1 */ }
        md, err = conv.ConvertPullRequest(pr)
    case parser.ResourceDiscussion:
        disc, err := client.FetchDiscussion(ctx, info.Owner, info.Repo, info.Number)
        if err != nil { /* stderr + exit 1 */ }
        md, err = conv.ConvertDiscussion(disc)
    }
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    if params.OutputFile != "" {
        // 自动创建父目录
        if err := os.MkdirAll(filepath.Dir(params.OutputFile), 0755); err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        if err := os.WriteFile(params.OutputFile, []byte(md), 0644); err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
    } else {
        fmt.Print(md)
    }
}
```

---

## 6. 实现阶段 (Phase Plan)

按 TDD 原则，每个阶段从写测试开始。

### Phase 1: 基础设施

| 步骤 | 文件 | 内容 |
|---|---|---|
| 1.1 | `Makefile` | `test`, `build`, `web`, `clean`, `fmt`, `vet` 命令 |
| 1.2 | `internal/parser/url_test.go` | 表格驱动测试：合法 URL × 6、非法 URL × 4、边界 case × 3 |
| 1.3 | `internal/parser/url.go` | `ParseURL()` 实现，正则 + `net/url` 解析 |
| 1.4 | `go get github.com/google/go-github/v69` | 安装依赖 |

**URL 测试用例示例**：

| input | expected |
|---|---|
| `https://github.com/owner/repo/issues/42` | issue, owner, repo, 42 |
| `https://github.com/owner/repo/pull/128` | pull, owner, repo, 128 |
| `https://github.com/owner/repo/discussions/99` | discussion, owner, repo, 99 |
| `https://github.com/owner/repo/issues/42/` | issue (尾部斜杠) |
| `https://github.com/owner/repo/issues/42?foo=bar` | issue (查询参数) |
| `not-a-url` | error |
| `https://github.com/owner/repo` | error (无类型) |

### Phase 2: GitHub 数据层 + 内部模型

| 步骤 | 文件 | 内容 |
|---|---|---|
| 2.1 | `internal/github/models.go` | 定义 Issue, PullRequest, Discussion, Comment, ReactionCount |
| 2.2 | `internal/github/mapping.go` | go-github → 内部模型映射函数 |
| 2.3 | `internal/github/client.go` | NewClient, 通用 getREST 方法 |
| 2.4 | `internal/github/client_test.go` | 集成测试：获取真实 golang/go Issue #1 |
| 2.5 | `internal/github/graphql.go` | GraphQL query 构造 + Discussion 获取 |

### Phase 3: Markdown 渲染器

| 步骤 | 文件 | 内容 |
|---|---|---|
| 3.1 | `internal/converter/converter_test.go` | 表格驱动测试：构造 Issue → 期望 Markdown |
| 3.2 | `internal/converter/frontmatter.go` | YAML Frontmatter 生成 |
| 3.3 | `internal/converter/comments.go` | 评论渲染（含缩进、Answer 标记、Review 标记） |
| 3.4 | `internal/converter/mentions.go` | @user → Markdown 链接 |
| 3.5 | `internal/converter/converter.go` | ConvertIssue/PR/Discussion 编排 |

### Phase 4: CLI 集成

| 步骤 | 文件 | 内容 |
|---|---|---|
| 4.1 | `internal/cli/parse_test.go` | 表格驱动测试：各种参数组合 |
| 4.2 | `internal/cli/parse.go` | flag 解析 |
| 4.3 | `internal/config/token.go` | GITHUB_TOKEN 读取 |
| 4.4 | `cmd/issue2md/main.go` | 主流程编排 |

### Phase 5: 端到端验证

| 步骤 | 内容 |
|---|---|
| 5.1 | `make build` 成功 |
| 5.2 | `make test` 全部通过 (单元 + 集成) |
| 5.3 | 手动测试: 转换公开 Issue / PR / Discussion |
| 5.4 | 验证 AC-01 到 AC-08、AC-E01 到 AC-E06 |

---

## 7. 构建与测试

### Makefile 命令

```makefile
.PHONY: build test test-unit test-integration clean fmt vet

build:
	go build -o bin/issue2md ./cmd/issue2md

test: test-unit test-integration

test-unit:
	go test -v -short ./internal/...

test-integration:
	go test -v ./internal/github/...  # 需要 GITHUB_TOKEN

clean:
	rm -rf bin/

fmt:
	go fmt ./...

vet:
	go vet ./...
```

### 测试命令设计说明

- `make test` 同时运行单元测试和集成测试
- `make test-unit` 通过 `-short` 标记跳过需要网络的集成测试（从 spec 继承）
- `make test-integration` 单独运行需要网络的 API 测试
- 集成测试在 `GITHUB_TOKEN` 未设置时应自动跳过（使用 `testing.Short()` 或 `t.Skip`）

### 集成测试跳过逻辑

```go
func TestFetchRealIssue(t *testing.T) {
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        t.Skip("GITHUB_TOKEN not set, skipping integration test")
    }
    // ... 测试逻辑
}
```

---

## 8. 注意事项与风险

| 风险 | 缓解措施 |
|---|---|
| Discussion GraphQL 响应结构复杂 | 提前在 GitHub API Explorer 中验证 query；使用 `map[string]any` + 类型断言逐步解析 |
| go-github 版本接口变更 | `go.mod` 锁定版本；`go.sum` 校验 |
| Rate limit 导致集成测试不稳定 | 集成测试使用 `t.Skip` 跳过当 token 耗尽时；提供 `-short` 模式 |
| YAML Frontmatter 中的特殊字符转义 | 对 Title 等字段中的 `"` 和 `:` 进行转义处理 |
