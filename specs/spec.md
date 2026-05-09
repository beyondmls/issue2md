# issue2md 产品需求规格说明书 (Spec)

> 版本: v0.1 (MVP)
> 状态: Draft
> 最后更新: 2026-05-09

---

## 1. 产品概述

issue2md 是一个命令行工具（CLI），输入一个 GitHub Issue / Pull Request / Discussion 的 URL，自动将其内容转换为结构化的 Markdown 文件，用于本地归档、文档化或知识管理。

### 1.1 核心原则

- **简单性优先**：不引入非必需的依赖，不做过度的抽象。
- **标准库优先**：优先使用 Go 标准库实现功能。
- **专注于公有仓库**：MVP 仅支持公有仓库内容读取。

---

## 2. 用户故事 (User Stories)

### MVP (v0.1)

| ID | 描述 | 优先级 |
|---|---|---|
| US-01 | 作为一名开发者，我想通过 CLI 命令传入一个 GitHub Issue URL，将其内容转为 Markdown 输出到终端 | P0 |
| US-02 | 作为一名开发者，我想通过 `--output` 参数将 Markdown 写入文件，而不是输出到 stdout | P0 |
| US-03 | 作为一名开发者，我想转换一个 GitHub Pull Request，获取其描述和 Review 评论（不含 Diff），用于归档讨论过程 | P0 |
| US-04 | 作为一名开发者，我想转换一个 GitHub Discussion，保留线程结构和所有评论 | P0 |
| US-05 | 作为一名开发者，我想通过 `GITHUB_TOKEN` 环境变量提供认证，以获得更高的 API 速率限制 | P1 |
| US-06 | 作为一名开发者，我想通过 `--with-reactions` 参数控制是否在输出中显示 Reactions 统计 | P1 |
| US-07 | 作为一名开发者，我想通过 `--with-user-links` 参数控制是否将用户名渲染为指向其 GitHub 主页的超链接 | P1 |

### 未来 (Post-MVP)

| ID | 描述 |
|---|---|
| US-F01 | 支持批量转换（从一个文件读取多个 URL） |
| US-F02 | 提供 Web 界面（Web UI），通过浏览器粘贴 URL 并下载 Markdown |
| US-F03 | 支持自定义输出模板 |
| US-F04 | 支持 GitLab / Gitee 等非 GitHub 平台 |
| US-F05 | 支持下离线缓存已获取的内容 |

---

## 3. 命令行界面 (CLI)

### 3.1 基本用法

```bash
# 输出到 stdout（默认）
issue2md https://github.com/owner/repo/issues/42

# 输出到文件
issue2md https://github.com/owner/repo/issues/42 -o ./docs/issue-42.md

# 输出 PR，附带 Reactions 和用户链接
issue2md https://github.com/owner/repo/pull/128 \
  --with-reactions \
  --with-user-links \
  -o ./docs/pr-128-review.md

# 输出 Discussion
issue2md https://github.com/owner/repo/discussions/99 -o ./docs/discussion-99.md
```

### 3.2 参数定义

| 参数 | 别名 | 类型 | 默认值 | 描述 | 关联用户故事 |
|---|---|---|---|---|---|
| `<url>` | — | String | 必填 | GitHub Issue/PR/Discussion URL | US-01, US-03, US-04 |
| `--output` | `-o` | String | `""` (stdout) | 输出文件路径 | US-02 |
| `--with-reactions` | — | Flag | `false` | 是否显示 Reactions 统计 | US-06 |
| `--with-user-links` | — | Flag | `false` | 是否将用户名渲染为 GitHub 链接 | US-07 |

**关于 Token 的说明**：
- 不提供 `--token` 参数。Token 仅通过环境变量 `GITHUB_TOKEN` 传入。
- 当 `GITHUB_TOKEN` 未设置时，以未认证模式运行（适用于公有仓库，速率限制较低）。

---

## 4. 功能规格

### 4.1 URL 解析

工具需要解析以下三类 GitHub URL 格式：

| 类型 | URL Pattern |
|---|---|
| Issue | `https://github.com/{owner}/{repo}/issues/{number}` |
| Pull Request | `https://github.com/{owner}/{repo}/pull/{number}` |
| Discussion | `https://github.com/{owner}/{repo}/discussions/{number}` |

**规则**：
- 输出 `owner`、`repo`、`type`（issue/pull/discussion）、`number`。
- 对于无法解析的 URL，返回清晰的错误信息并退出（exit code 1）。

### 4.2 数据获取

#### 4.2.1 GitHub API 路由

| 资源 | API | 端点 |
|---|---|---|
| Issue 详情 | REST v3 | `GET /repos/{owner}/{repo}/issues/{number}` |
| Issue 评论 | REST v3 | `GET /repos/{owner}/{repo}/issues/{number}/comments` |
| PR 详情 | REST v3 | `GET /repos/{owner}/{repo}/pulls/{number}` |
| PR Review 评论 | REST v3 | `GET /repos/{owner}/{repo}/pulls/{number}/comments` |
| Discussion | GraphQL v4 | 自定义 query |
| Discussion 评论 | GraphQL v4 | 自定义 query (通过 `comments` 连接) |

**注意**：
- Discussions 仅可通过 GitHub GraphQL API (v4) 访问。因此 MVP 需要同时支持 REST v3 和 GraphQL v4。
- PR 的常规 Issue 评论通过 `GET /repos/{owner}/{repo}/issues/{number}/comments` 获取。
- PR 的 Review 评论通过 `GET /repos/{owner}/{repo}/pulls/{number}/comments` 获取。两者都需要输出。

#### 4.2.2 错误处理

| 场景 | 行为 |
|---|---|
| 无效 URL | 报错 "invalid GitHub URL: ..."，exit code 1 |
| 网络超时 / 连接失败 | 报错 "failed to fetch: ..."，exit code 1 |
| HTTP 403 (Rate Limit) | 提示用户设置 `GITHUB_TOKEN`，exit code 1 |
| HTTP 404 | 报错 "resource not found: ..."，exit code 1 |
| 非 2xx 响应 | 报错 "unexpected status: <code> <body>"，exit code 1 |
| 输出文件路径不存在 | 自动创建目录（类似 `mkdir -p` 行为） |
| 输出文件已存在 | 覆写，不提示（CLI 工具惯例） |

### 4.3 Markdown 输出格式

#### 4.3.1 整体结构

```markdown
# Title (#{number})

> **Author**: [username](https://github.com/username) (当 `--with-user-links` 开启)
> **State**: open | closed
> **Created**: 2026-05-09T12:00:00Z
> **Labels**: bug, enhancement
> **Reactions**: 👍 5 🎉 2 ❤️ 3 (当 `--with-reactions` 开启)

---

## Description

<!-- Issue/PR/Discussion 的原始正文内容，按 GitHub Flavored Markdown 输出 -->

---

## Comments

### Comment by **username** — 2026-05-09T13:00:00Z

Comment body in GitHub Flavored Markdown...

> **Reactions**: 👍 2 🎉 1 (当 `--with-reactions` 开启)

---

### Comment by **another-user** — 2026-05-09T14:00:00Z

Another comment...

> **Reactions**: ❤️ 3 (当 `--with-reactions` 开启)
```

#### 4.3.2 差异处理

**Pull Request**：
- 标题为 "Title (#123)"
- "Description" 部分包含 PR 的 body 内容
- 不含 Diff / Commits / Changed Files
- Comments 部分包含：
  - Pull Request 上的常规 Issue 评论
  - Review 评论（包括每个 Review 下的子评论，以缩进或分隔线区分）

**Discussion**：
- 标题为 "Title (#99)"
- "Description" 部分是 Discussion 的 body 内容
- 保留评论的层级关系（回复/子线程 → 缩进或分隔线）
- 保留 Answer 标记（如果有被标记为回答的评论，需突出显示）

#### 4.3.3 内容渲染规则

| 元素 | 处理方式 |
|---|---|
| `@mentions` | 当 `--with-user-links` 开启时，渲染为链接；否则保持原样 |
| `#123` 引用 | 直接保留，不做转换 |
| 代码块 | 原样保留 |
| 图片链接 | 原样保留，不做下载 |
| 任务列表 (`- [ ]`) | 原样保留 |
| Emoji | 原样保留 |
| 表格 | 原样保留 |

---

## 5. 技术架构

### 5.1 包结构

```
.
├── cmd/
│   └── issue2md/
│       └── main.go              # CLI 入口，flag 解析，调用流程编排
├── internal/
│   ├── github/
│   │   ├── client.go            # GitHub API 客户端（REST + GraphQL）
│   │   ├── client_test.go       # 集成测试
│   │   ├── issue.go             # Issue 类型定义 & 获取逻辑
│   │   ├── pull.go              # PR 类型定义 & 获取逻辑
│   │   ├── discussion.go        # Discussion 类型定义 & 获取逻辑（GraphQL）
│   │   └── url.go               # URL 解析
│   ├── markdown/
│   │   ├── renderer.go          # Markdown 渲染器
│   │   ├── renderer_test.go     # 单元测试
│   │   └── templates.go         # 模板常量（如果后续需要用 template 的话）
│   └── config/
│       └── config.go            # 配置解析（环境变量等）
├── specs/
│   └── spec.md                  # 本文件
├── Makefile
├── go.mod
├── CLAUDE.md
└── constitution.md
```

### 5.2 依赖策略

按宪法要求，**优先使用标准库**：

| 需求 | 方案 |
|---|---|
| HTTP 请求 | `net/http`（标准库） |
| JSON 解析 | `encoding/json`（标准库） |
| CLI flag 解析 | `flag`（标准库） |
| GraphQL 查询 | `net/http` + 手动构造 JSON body（标准库） |
| 字符串模板 | `fmt` / `strings.Builder`（标准库） |
| 第三方依赖 | **零外部依赖**（MVP 阶段） |

### 5.3 数据流

```
URL
  │
  ▼
URL Parser ──► {owner, repo, type, number}
  │
  ▼
GitHub API Client ──► GitHub REST / GraphQL
  │
  ▼
Internal Models ──► Issue / PullRequest / Discussion structs
  │
  ▼
Markdown Renderer ──► string (Markdown)
  │
  ▼
Output ──► stdout / file
```

---

## 6. 非功能需求

### 6.1 性能

- 发起 2-4 个并发 API 请求来获取资源和评论（而非串行），以降低延迟。
- 单次转换应在 5 秒内完成（网络正常条件下）。

### 6.2 错误可观测性

- 所有错误使用 `fmt.Errorf("...: %w", err)` 包装上下文。
- Exit code 语义：
  - `0`：成功
  - `1`：可预期的错误（无效 URL、网络错误、API 拒绝等）

### 6.3 安全性

- Token 仅通过环境变量传入，绝不在日志或错误信息中泄露。
- 不对用户 URL 做任何 shell 注入风险的操作。

---

## 7. 测试策略

遵循宪法第二条（测试先行铁律）：

| 层级 | 方法 | 范围 |
|---|---|---|
| 单元测试 | 表格驱动测试 | URL 解析、Markdown 渲染 |
| 集成测试 | 使用真实的 `github.com` | GitHub API 客户端（通过 `GITHUB_TOKEN`） |
| 端到端测试 | Shell 脚本 | 完整的 CLI 调用流程 |

---

## 8. 未来展望（Post-MVP）

以下功能已明确知道但不纳入 v0.1：

1. **Web 界面 (Web UI)**：提供一个简单的 Web 服务，用户通过浏览器粘贴 URL 即可获取 Markdown 文件。
2. **批量转换**：支持从文件读取多个 URL，批量转换为多个 Markdown 文件。
3. **自定义模板**：允许用户通过 Go template 自定义输出格式。
4. **多平台支持**：支持 GitLab、Gitee 等平台的 Issue 转换为 Markdown。
