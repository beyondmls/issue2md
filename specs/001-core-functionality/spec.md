# issue2md — Core Functionality Spec

> **版本**: v0.1 (MVP)
> **状态**: Final Draft
> **最后更新**: 2026-05-09

---

## 1. 产品概述

issue2md 是一个命令行工具（CLI），输入一个 GitHub Issue / Pull Request / Discussion 的 URL，自动将其内容转换为带 YAML Frontmatter 的结构化 Markdown 文件，用于本地归档、文档化或知识管理。

---

## 2. 用户故事

### 2.1 MVP 用户故事 (v0.1 — CLI)

| ID | 描述 | 优先级 |
|---|---|---|
| US-01 | 作为开发者，我想通过一条 CLI 命令传入任意 GitHub Issue/PR/Discussion 的 URL，工具自动识别类型并输出 Markdown，无需手动指定类型 | P0 |
| US-02 | 作为开发者，我可以选择将 Markdown 输出到文件或 stdout，灵活适配管道和文件写入场景 | P0 |
| US-03 | 作为开发者，我希望转换 PR 时能保留描述和所有 Review 评论（不含 Diff），用于归档完整的讨论过程 | P0 |
| US-04 | 作为开发者，我希望转换 Discussion 时能保留所有评论及其层级关系，并标记被选为 Answer 的评论 | P0 |
| US-05 | 作为开发者，我可以设置 `GITHUB_TOKEN` 环境变量来获得更高的 API 速率限制 | P1 |
| US-06 | 作为开发者，我可以通过 `-enable-reactions` 标志控制是否在输出中包含 Reactions 统计 | P1 |
| US-07 | 作为开发者，我可以通过 `-enable-user-links` 标志控制是否将用户名渲染为指向其 GitHub 主页的超链接 | P1 |

### 2.2 未来用户故事 (Post-MVP)

| ID | 描述 |
|---|---|
| US-F01 | 作为开发者，我想提供一个 URL 列表文件，批量转换为多个 Markdown 文件 |
| US-F02 | 作为用户，我想通过浏览器访问一个 Web 页面，粘贴 URL 即可下载 Markdown |
| US-F03 | 作为开发者，我希望使用自定义 Go Template 控制输出格式 |
| US-F04 | 作为开发者，我想支持 GitLab / Gitee 等非 GitHub 平台 |
| US-F05 | 作为开发者，我希望支持离线缓存以加速重复检索 |

---

## 3. 功能性需求

### 3.1 命令行接口

**原型**:

```
issue2md [flags] <url> [output_file]
```

**位置参数**:

| 参数 | 必填 | 描述 |
|---|---|---|
| `<url>` | 是 | GitHub Issue/PR/Discussion 的完整 URL |
| `[output_file]` | 否 | 输出文件路径。不提供时输出到 stdout |

**标志 (Flags)**:

| 标志 | 类型 | 默认 | 描述 |
|---|---|---|---|
| `-enable-reactions` | bool | `false` | 在每个主帖和评论下方添加 Reactions 统计行 |
| `-enable-user-links` | bool | `false` | 将 `@username` 渲染为指向 `https://github.com/username` 的 Markdown 链接 |

**环境变量**:

| 变量 | 必填 | 描述 |
|---|---|---|
| `GITHUB_TOKEN` | 否 | GitHub Personal Access Token。未设置时以未认证模式运行 |

**Exit Codes**:

| Code | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 可预期的错误（无效 URL、网络错误、API 拒绝、资源不存在等） |

**Token 安全**:

- Token **仅**通过环境变量 `GITHUB_TOKEN` 获取。
- **不提供** `-token` 命令行标志，以防 Token 泄露到 Shell 历史记录中。

### 3.2 URL 自动识别

工具必须自动解析 URL 结构来决定资源类型，无需用户指定。必须处理以下格式：

```
https://github.com/{owner}/{repo}/issues/{number}
https://github.com/{owner}/{repo}/pull/{number}
https://github.com/{owner}/{repo}/discussions/{number}
```

URL 允许以尾部 `/` 结尾（如 `/issues/42/`），允许包含无关查询参数（如 `?foo=bar`）。

**解析输出**: `{owner, repo, type: issue|pull|discussion, number}`

### 3.3 GitHub API 路由

| 资源类型 | API 版本 | 端点 |
|---|---|---|
| Issue 详情 | REST v3 | `GET /repos/{owner}/{repo}/issues/{number}` |
| Issue 评论 | REST v3 | `GET /repos/{owner}/{repo}/issues/{number}/comments` |
| PR 详情 | REST v3 | `GET /repos/{owner}/{repo}/pulls/{number}` |
| PR Review 评论 | REST v3 | `GET /repos/{owner}/{repo}/pulls/{number}/comments` |
| PR 常规评论 (同 Issue 评论) | REST v3 | `GET /repos/{owner}/{repo}/issues/{number}/comments` |
| Discussion 详情 + 评论 | GraphQL v4 | 自定义 GraphQL Query |

**关于 PR 的评论输出**:

- PR 上的常规 Issue 评论（`/issues/{number}/comments`）和 Review 评论（`/pulls/{number}/comments`）**合并后按时间正序平铺展示**。
- **不需要**按 Review 分组。目标是展示"发生了什么对话"，而不是代码审核的结构。
- 每条评论标注 **类型提示**：`[Review Comment]` 或 `[Comment]`。

**关于 Discussion**:

- Discussions 仅能通过 GraphQL API (v4) 访问。
- 必须保留评论的线程层级关系（父/子），通过缩进或分隔线体现。
- 如果某条评论被标记为 **Answer**（Accepted Answer），必须在渲染时给予显著标记。

### 3.4 排序规则

- 所有评论（Issue 评论、PR 评论、Review 评论、Discussion 回复）**统一按创建时间正序排列**。
- 第一条始终是 Issue/PR/Discussion 本身的"主楼"内容（Description）。
- 之后按时间顺序列出所有评论。

### 3.5 评论内容范围

| 元素 | 处理方式 |
|---|---|
| Markdown 正文 | 原样保留（GitHub Flavored Markdown） |
| `@mentions` | 原样保留。当 `-enable-user-links` 开启时，渲染为 Markdown 链接 |
| `#123` 引用 | 保留纯文本。不做转换 |
| 代码块 (含语言标记) | 原样保留 |
| 图片链接 | 原样保留。**不下载**到本地 |
| 任务列表 `- [ ]` | 原样保留 |
| Emoji | 原样保留 |
| 表格 | 原样保留 |
| 数学公式 (`$...$`) | 原样保留 |
| Reactions | 仅当 `-enable-reactions` 开启时渲染。格式: `👍 5 🎉 2 ❤️ 3` |

### 3.6 输出格式 — Markdown 结构

#### 3.6.1 YAML Frontmatter

每个输出文件头部包含 YAML Frontmatter，内容示例：

```yaml
---
title: "Fix login timeout issue"
url: https://github.com/owner/repo/issues/42
number: 42
type: issue
author: octocat
state: open
created_at: "2026-04-15T10:30:00Z"
updated_at: "2026-04-16T14:20:00Z"
labels:
  - bug
  - performance
milestone: v2.0
---
```

Frontmatter 字段：

| 字段 | 必填 | 说明 |
|---|---|---|
| `title` | 是 | 标题（不含 `#编号` 后缀） |
| `url` | 是 | 原始 GitHub URL |
| `number` | 是 | Issue/PR/Discussion 编号 (数值) |
| `type` | 是 | `issue` / `pull` / `discussion` |
| `author` | 是 | 创建者的 GitHub 用户名 |
| `state` | 是 | `open` / `closed` |
| `created_at` | 是 | ISO 8601 格式的创建时间 |
| `updated_at` | 是 | ISO 8601 格式的最后更新时间 |
| `labels` | 否 | 标签列表（Issue 和 PR 有，Discussion 可能没有） |
| `milestone` | 否 | 里程碑名称（可能有） |

#### 3.6.2 正文结构

**Standard Markdown 正文**（以 YAML Frontmatter 后的空行分隔）：

```markdown
---
title: "..."
url: ...
---

# Title (#42)

<!-- description body in GFM -->

---

## Comments

### [Comment] by **user1** — 2026-04-15T11:00:00Z

Comment body text...

> Reactions: 👍 5 🎉 2

---

### [Comment] by **user2** — 2026-04-15T12:00:00Z

Another comment...
```

**Discussion Answer 标记**：当 `-enable-user-links` 开启时 `@user` 转为链接。

被标记为 Answer 的 Discussion 评论，额外渲染为：

```markdown
### ✅ Answer by **user1** — 2026-04-15T11:00:00Z

> Answer body...
```

**Discussion 子评论缩进**：子评论（回复）在父评论下方，通过 `> ` 引用块或 `&emsp;` 缩进 + 标注 `[Reply]`。

**PR Review Comments 标记**：

```markdown
### [Review Comment] by **reviewer** on `path/to/file` (line 42) — 2026-04-15T11:00:00Z

Comment body...
```

Review Comment 的 `path` 和 `line` 信息来自 GitHub API 的 `path` 和 `line` / `position` 字段。当无法确定行号时，不渲染 `(line N)` 部分。

#### 3.6.3 分隔线

- Frontmatter 和正文之间：标准 YAML 分隔（`---` + 空行）
- 主楼（Description）和 Comments 区域之间：`---`
- 每条评论之间：`---`

---

## 4. 非功能性需求

### 4.1 架构原则

- **无全局变量**: 所有依赖通过结构体成员和函数参数显式注入。
- **标准库优先**: CLI 框架使用 `flag` 标准库，HTTP 使用 `net/http`，JSON 使用 `encoding/json`。
- **零外部依赖**: MVP 阶段不引入第三方 Go 模块。
- **解耦设计**: `internal/github/` (API 客户端) 与 `internal/markdown/` (渲染器) 之间通过内部 Model 层解耦，互不依赖。

### 4.2 数据流

```
URL
  │ (cmd/issue2md/main.go)
  ▼
URL Parser ──► {owner, repo, type, number}
  │ (internal/github/url.go)
  ▼
GitHub API Client ──► REST / GraphQL
  │ (internal/github/client.go + issue.go / pull.go / discussion.go)
  ▼
Internal Models ──► Issue / PullRequest / Discussion structs
  │ (internal/github/models.go)
  ▼
Markdown Renderer ──► string (Markdown with YAML frontmatter)
  │ (internal/markdown/renderer.go)
  ▼
Output ──► stdout / file
  │ (cmd/issue2md/main.go)
```

### 4.3 错误处理与可观测性

- 所有错误必须使用 `fmt.Errorf("...: %w", err)` 包装上下文。
- 错误信息输出到 **stderr**，不污染 stdout 的 Markdown 内容。
- **无重试机制**：网络错误或 API 错误直接透传给用户。
- Rate limit 错误时，透传 GitHub API 返回的错误信息（包括 `X-RateLimit-Reset` 头部信息）。
- 程序退出码严格遵循 `0` = 成功，`1` = 失败。

### 4.4 性能目标

- 单次 URL 转换在正常网络条件下应在 5 秒内完成。
- 对于需要多个 API 请求的资源（如 PR 需要获取 2-3 个端点），使用并发请求以减少等待时间。

### 4.5 安全性

- `GITHUB_TOKEN` 绝不在日志、错误提示或进程列表中泄露。
- 不对用户输入的 URL 做 shell 注入风险的操作。
- 不对 URL 做任何额外网络请求之外的副作用操作。

---

## 5. 验收标准 (Acceptance Criteria)

### 5.1 功能验收

| # | 场景 | 输入 | 预期输出 |
|---|---|---|---|
| AC-01 | 转换一个公开 Issue | `issue2md https://github.com/golang/go/issues/1` | 正确输出 YAML Frontmatter + Markdown 正文到 stdout |
| AC-02 | 转换并写入文件 | `issue2md https://github.com/golang/go/issues/1 /tmp/go-issue-1.md` | 文件 `/tmp/go-issue-1.md` 内容正确 |
| AC-03 | 转换公开 PR | `issue2md https://github.com/golang/go/pull/100` | 输出包含 PR body + 所有 Review 评论，按时间正序 |
| AC-04 | 转换公开 Discussion | `issue2md https://github.com/community/community/discussions/1` | 输出包含 YAML Frontmatter、所有评论层级、Answer 标记 |
| AC-05 | 自动识别资源类型 | Issue URL / PR URL / Discussion URL | 正确识别并路由到对应的 API 处理逻辑 |
| AC-06 | Reactions 标志 | `issue2md -enable-reactions <url>` | 输出中包含 Reactions 统计行 |
| AC-07 | User links 标志 | `issue2md -enable-user-links <url>` | 输出中 `@user` 渲染为 Markdown 链接 |
| AC-08 | 组合 Flags | `issue2md -enable-reactions -enable-user-links <url>` | 同时启用两个特性 |

### 5.2 错误场景验收

| # | 场景 | 输入 | 预期输出 |
|---|---|---|---|
| AC-E01 | 无效 URL | `issue2md not-a-url` | stderr 输出错误信息，exit code 1 |
| AC-E02 | 不支持的 URL | `issue2md https://github.com/golang/go` | stderr 输出错误信息，exit code 1 |
| AC-E03 | 不存在的资源 | `issue2md https://github.com/golang/go/issues/99999999` | stderr 输出 404 错误，exit code 1 |
| AC-E04 | 输出到不存在的目录 | `issue2md <url> /nonexistent/dir/file.md` | 自动创建目录并写入成功 |
| AC-E05 | 无 token 时访问 rate-limited API | 未设 `GITHUB_TOKEN`，高频请求 | stderr 输出 rate limit 错误，exit code 1 |
| AC-E06 | 网络不可达 | `issue2md https://github.invalid/test/test/issues/1` | stderr 输出网络错误，exit code 1 |

### 5.3 输出格式验收

| # | 场景 | 预期 |
|---|---|---|
| AC-F01 | Frontmatter 包含所有必填字段 | `title`, `url`, `number`, `type`, `author`, `state`, `created_at`, `updated_at` |
| AC-F02 | 评论按时间正序排列 | 第一条评论是最早的，最后一条是最新的 |
| AC-F03 | Discussion Answer 有特殊标记 | 被标记为 Answer 的评论显示 `✅ Answer` 及引用块 |
| AC-F04 | PR Review 评论标记 | 每条 Review 评论显示 `[Review Comment]` 及文件路径 |
| AC-F05 | Markdown 内容完整 | 代码块、列表、表格、链接、图片等均原样保留 |

### 5.4 测试策略

遵循宪法第二条（测试先行铁律）：

| 层级 | 技术 | 场景覆盖 |
|---|---|---|
| 单元测试 | 表格驱动测试 | URL 解析: 各种合法/非法 URL 格式 |
| 单元测试 | 表格驱动测试 | Markdown 渲染: 各种内部模型构造、Frontmatter 生成 |
| 单元测试 | 表格驱动测试 | Markdown 渲染: Flags 对输出的影响 |
| 集成测试 | 使用真实 `github.com` API + `GITHUB_TOKEN` | 获取真实公开 Issue/PR/Discussion |
| 端到端测试 | Shell 脚本调用二进制 | 完整的 CLI 调用、exit code、stdout/stderr 验证 |

---

## 6. 输出格式完整示例

### 6.1 Issue 示例

```markdown
---
title: "encoding/json: add Decoder.Decode method"
url: https://github.com/golang/go/issues/100
number: 100
type: issue
author: rsc
state: closed
created_at: "2026-04-01T09:00:00Z"
updated_at: "2026-04-10T17:30:00Z"
labels:
  - proposal
  - v2
milestone: Go2
---

# encoding/json: add Decoder.Decode method (#100)

We should consider adding a Decode method to json.Decoder...

---

## Comments

### [Comment] by **gopherbot** — 2026-04-01T10:00:00Z

This has been discussed before in #42.

---

### [Comment] by **rsc** — 2026-04-05T14:00:00Z

I've thought about this more and I think it's the right approach.
```

### 6.2 Discussion 示例（含 Answer）

```markdown
---
title: "Best practices for Go error handling in 2026"
url: https://github.com/community/community/discussions/42
number: 42
type: discussion
author: gopher
state: open
created_at: "2026-04-01T09:00:00Z"
updated_at: "2026-04-10T17:30:00Z"
---

# Best practices for Go error handling in 2026 (#42)

I'm wondering what the community thinks about error handling...

---

## Comments

### ✅ Answer by **davecheney** — 2026-04-02T10:00:00Z

> The current best practice is to use `fmt.Errorf` with `%w` for wrapping...

---

### [Comment] by **robpike** — 2026-04-03T11:00:00Z

I agree with Dave's assessment. Additionally...
```

### 6.3 PR 示例

```markdown
---
title: "Fix race condition in http.Server"
url: https://github.com/golang/go/pull/500
number: 500
type: pull
author: bradfitz
state: open
created_at: "2026-05-01T09:00:00Z"
updated_at: "2026-05-02T17:30:00Z"
labels:
  - security
---

# Fix race condition in http.Server (#500)

This PR fixes a race condition in the HTTP server shutdown path...

---

## Comments

### [Comment] by **iansmith** — 2026-05-01T12:00:00Z

This looks reasonable, but have you considered the timeout case?

---

### [Review Comment] by **dvyukov** on `net/http/server.go` (line 342) — 2026-05-01T14:00:00Z

There's a similar race on this line. `srv.active` isn't protected here.

---

### [Comment] by **bradfitz** — 2026-05-01T15:00:00Z

Good catch, adding a mutex there too.
```

---

## 7. 技术约束

- **语言**: Go 1.26+（以 `go.mod` 为准）
- **外部依赖**: **零外部依赖**（仅使用 Go 标准库）
- **构建工具**: 通过 `Makefile` 管理构建和测试命令
- **代码风格**: 遵循 Go 官方 `gofmt` / `govet` 规范
- **测试框架**: 内置 `testing` 包 + 表格驱动测试风格
- **GraphQL 查询**: 通过 `net/http` 发送 POST 请求，手动构造 JSON body，不做深度抽象
