# issue2md — Task List

> **基于**: `spec.md` + `plan.md`
> **阶段**: v0.1 (MVP)
> **约定**: `[P]` = 可并行执行的任务；TDD = 测试先于实现

---

## Phase 1: Foundation (数据结构定义)

> 目标：搭建项目骨架，定义核心模型和工具函数。
> **出口条件**: `make test-unit` 全部通过，parser/cli/config 三层覆盖。

### Task 1.1 — 创建 Makefile

| 字段 | 值 |
|---|---|
| 文件 | `Makefile` |
| 内容 | `build`, `test`, `test-unit`, `test-integration`, `clean`, `fmt`, `vet` 目标 |
| 并行 | `[P]` |
| 依赖 | 无 |
| TDD | 否 — 基础设施，无测试 |

### Task 1.2 — 添加 go-github 依赖

| 字段 | 值 |
|---|---|
| 文件 | `go.mod` + `go.sum` |
| 命令 | `go get github.com/google/go-github/v69` |
| 并行 | `[P]` |
| 依赖 | Task 1.1 (Makefile 需包含 build 目标) |
| TDD | 否 |

### Task 1.3 — 定义内部数据模型

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/models.go` |
| 内容 | `Issue`, `PullRequest`, `Discussion`, `Comment`, `ReactionCount`, `GitHubResource` — 按 plan.md §4 定义 |
| 并行 | `[P]` |
| 依赖 | 无 |
| TDD | 否 — 纯 struct 定义，无行为逻辑 |

### Task 1.4 — 编写 URL 解析器单元测试

| 字段 | 值 |
|---|---|
| 文件 | `internal/parser/url_test.go` |
| 内容 | 表格驱动测试: 合法 URL × 6 (含尾部斜杠/查询参数)、非法 URL × 4、边界 case × 3 |
| 并行 | `[P]` |
| 依赖 | 无 |
| TDD | **Red** — 先写断言，预期编译失败 |

### Task 1.5 — 实现 URL 解析器

| 字段 | 值 |
|---|---|
| 文件 | `internal/parser/url.go` + `internal/parser/errors.go` |
| 内容 | `ParseURL()`, `URLInfo`, `ResourceType`, `ErrInvalidURL` |
| 依赖 | Task 1.4 (TDD: 测试先行) |
| TDD | **Green** — 让 Task 1.4 的测试通过 |

### Task 1.6 — 编写 CLI 参数解析单元测试

| 字段 | 值 |
|---|---|
| 文件 | `internal/cli/parse_test.go` |
| 内容 | 表格驱动测试: 默认值、`-enable-reactions`、`-enable-user-links`、输出文件位置参数、参数缺失、无效 flag |
| 并行 | `[P]` |
| 依赖 | 无 |
| TDD | **Red** |

### Task 1.7 — 实现 CLI 参数解析器

| 字段 | 值 |
|---|---|
| 文件 | `internal/cli/parse.go` |
| 内容 | `Params` struct, `Parse()` — 基于 `flag` 标准库 |
| 依赖 | Task 1.6 (TDD) |
| TDD | **Green** |

### Task 1.8 — 编写配置读取单元测试

| 字段 | 值 |
|---|---|
| 文件 | `internal/config/token_test.go` |
| 内容 | `GITHUB_TOKEN` 设置/未设置两种场景 |
| 并行 | `[P]` |
| 依赖 | 无 |
| TDD | **Red** |

### Task 1.9 — 实现配置读取

| 字段 | 值 |
|---|---|
| 文件 | `internal/config/token.go` |
| 内容 | `GetGitHubToken()` — 仅 `os.Getenv` |
| 依赖 | Task 1.8 (TDD) |
| TDD | **Green** |

---

## Phase 2: GitHub Fetcher (API 交互逻辑)

> 目标：实现 GitHub API 客户端，能获取 Issue/PR/Discussion。
> **出口条件**: `make test-integration` 能成功获取 `golang/go#1` 等真实数据。

### Task 2.1 — 实现 go-github 类型映射

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/mapping.go` |
| 内容 | `mapIssue()`, `mapPullRequest()`, `mapComment()` — 将 `github.Issue` / `github.PullRequest` / `github.Comment` 映射为内部模型 |
| 并行 | `[P]` (可和 Task 1.4-1.9 并行) |
| 依赖 | Task 1.2 (go-github), Task 1.3 (models) |
| TDD | 否 — 纯映射逻辑，通过后续集成测试覆盖 |

### Task 2.2 — 编写 GitHub 客户端集成测试 (REST)

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/client_test.go` |
| 内容 | 集成测试: `TestFetchIssue_Real`, `TestFetchPullRequest_Real`；使用真实 `github.com`；无 `GITHUB_TOKEN` 时 `t.Skip` |
| 依赖 | Task 1.2 (go-github), Task 1.3 (models), Task 2.1 (mapping) |
| TDD | **Red** — 先写断言，预期集成测试因 client 未实现而失败 |

### Task 2.3 — 实现 GitHub 客户端 (Issue + PR)

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/client.go` |
| 内容 | `NewClient()`, `FetchIssue()`, `FetchPullRequest()` — 基于 go-github；`FetchPullRequest` 内部并发获取 review comments + issue comments 后合并排序 |
| 依赖 | Task 2.2 (TDD) |
| TDD | **Green** |

### Task 2.4 — 编写 GraphQL Discussion 集成测试

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/graphql_test.go` |
| 内容 | 集成测试: `TestFetchDiscussion_Real`；使用真实 Discussion URL；保留 Answer/层级断言 |
| 依赖 | Task 1.3 (models) |
| 并行 | `[P]` (和 Task 2.2 并行，都只需 models 即可写断言) |
| TDD | **Red** |

### Task 2.5 — 实现 GraphQL Discussion 获取

| 字段 | 值 |
|---|---|
| 文件 | `internal/github/graphql.go` |
| 内容 | GraphQL query 定义 + `FetchDiscussion()` 实现；解析嵌套响应为 `Discussion` + `Comment` (含 `IsAnswer`, `ParentID`) |
| 依赖 | Task 2.4 (TDD) |
| TDD | **Green** |

---

## Phase 3: Markdown Converter (转换逻辑)

> 目标：将内部数据模型渲染为带 YAML Frontmatter 的 Markdown。
> **出口条件**: `make test-unit` 覆盖所有 Frontmatter/评论/链接渲染场景。

### Task 3.1 — 编写转换器单元测试

| 字段 | 值 |
|---|---|
| 文件 | `internal/converter/converter_test.go` |
| 内容 | 表格驱动测试: Issue → Markdown、PR → Markdown (含 Review 标记)、Discussion → Markdown (含 Answer/层级)；Flags 开关测试 |
| 并行 | `[P]` (可和 Phase 2 并行，只需 models 即可构造输入) |
| 依赖 | Task 1.3 (models) |
| TDD | **Red** |

### Task 3.2 — 实现 YAML Frontmatter 生成

| 字段 | 值 |
|---|---|
| 文件 | `internal/converter/frontmatter.go` |
| 内容 | `renderFrontmatter()` — 手动 YAML 字符串拼接；处理 `"` / `:` 转义 |
| 依赖 | Task 3.1 (TDD) |
| TDD | **Green** (作为 converter.go 的子模块) |

### Task 3.3 — 实现评论渲染

| 字段 | 值 |
|---|---|
| 文件 | `internal/converter/comments.go` |
| 内容 | `renderComments()`, `renderComment()`, `renderReactions()` — 评论列表遍历；Answer 标记 `✅ Answer`；Review 标记 `[Review Comment]`；子评论缩进 |
| 依赖 | Task 3.1 (TDD) |
| TDD | **Green** (作为 converter.go 的子模块) |

### Task 3.4 — 实现 Mentions 链接渲染

| 字段 | 值 |
|---|---|
| 文件 | `internal/converter/mentions.go` |
| 内容 | `renderUserLinks()` — 正则替换 `@username` → `[@username](https://github.com/username)` |
| 依赖 | Task 3.1 (TDD) |
| TDD | **Green** (作为 converter.go 的子模块) |

### Task 3.5 — 实现转换器主流程

| 字段 | 值 |
|---|---|
| 文件 | `internal/converter/converter.go` |
| 内容 | `New()`, `Converter` struct, `ConvertIssue()`, `ConvertPullRequest()`, `ConvertDiscussion()` — 编排 frontmatter + comments + mentions |
| 依赖 | Task 3.2, 3.3, 3.4 (frontmatter/comments/mentions 就绪) |
| TDD | **Green** — 让 Task 3.1 的全部测试通过 |

---

## Phase 4: CLI Assembly (命令行入口集成)

> 目标：组装所有包，实现完整的命令行工具。
> **出口条件**: `make build` + `make test` 全部通过 + 端到端验证。

### Task 4.1 — 实现 CLI 主入口

| 字段 | 值 |
|---|---|
| 文件 | `cmd/issue2md/main.go` |
| 内容 | `main()` 编排: cli.Parse → config.GetToken → parser.ParseURL → github.Client → converter.Convert → stdout/文件输出 |
| 依赖 | Task 1.5 (parser), Task 1.7 (cli), Task 1.9 (config), Task 2.3 (client REST), Task 2.5 (graphql), Task 3.5 (converter) |
| TDD | 否 — 集成验证，通过 `make build` + 手动 E2E 测试确认 |

### Task 4.2 — 端到端验证

| 字段 | 值 |
|---|---|
| 操作 | 手动执行 E2E 测试 |
| 验证项 | `make build` 成功 → `make test` 全绿 → 对照 AC-01 到 AC-08 逐条验证 → 对照 AC-E01 到 AC-E06 逐条验证 |
| 依赖 | Task 4.1 |

---

## 依赖关系全景图

```
Phase 1                              Phase 2                        Phase 3                Phase 4
──────────────────────────────────────────────────────────────────────────────────────────────────
1.1 Makefile [P]
 │
1.2 go-github [P] ────────────┐
 │                           │
1.3 models.go [P] ───────┐   │   ┌─────────────────────────┐
 │                       │   │   │                         │
1.4 parser test [P]      │   │   │                         │
 │                       │   │   │                         │
1.5 parser impl ←────────┤   │   │                         │
 │                       │   │   │                         │
1.6 cli test [P]         │   │   │                         │
 │                       │   │   │                         │
1.7 cli impl ←───────────┤   │   │                         │
 │                       │   │   │                         │
1.8 config test [P]      │   │   │                         │
 │                       │   │   │                         │
1.9 config impl ←────────┘   │   │                         │
                             │   │                         │
                   2.1 mapping.go ←┘───────────────────┐   │
                             │                         │   │
                   2.2 client test ←────────────────┐  │   │
                             │                      │  │   │
                   2.3 client impl ←────────────────┘  │   │
                             │                         │   │
                   2.4 graphql test [P] ←──────────────┘   │
                             │                             │
                   2.5 graphql impl ←──────────────────────┘
                                                           │
                                                 3.1 converter test [P] ←──┘
                                                           │
                                                 3.2 frontmatter.go ←──────┤
                                                 3.3 comments.go ←────────┤
                                                 3.4 mentions.go ←────────┤
                                                           │
                                                 3.5 converter.go ←───────┘
                                                           │
                                                           └──────────┐
                                                                      │
                                                             4.1 main.go
                                                                      │
                                                             4.2 E2E verify
```

---

## 统计

| 指标 | 数量 |
|---|---|
| 总任务数 | 20 |
| 可并行任务 `[P]` | 10 (50%) |
| TDD 任务对 (Red → Green) | 6 对 (Task 1.4-1.5, 1.6-1.7, 1.8-1.9, 2.2-2.3, 2.4-2.5, 3.1→3.5) |
| 非 TDD 任务 | 4 (Makefile, go-github, models, mapping, main.go, E2E) |
