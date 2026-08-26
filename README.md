# zentao-cli

<div align="center">

<h3>🚀 面向 AI Agent 与现代开发者的禅道 (ZenTao PMS) CLI 工具与 Go SDK</h3>

[![Go Reference](https://pkg.go.dev/badge/github.com/windosx/zentao-cli/pkg/zentao.svg)](https://pkg.go.dev/github.com/windosx/zentao-cli/pkg/zentao)
[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/windosx/zentao-cli?include_prereleases&style=flat-square)](https://github.com/windosx/zentao-cli/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Goreleaser](https://img.shields.io/badge/release-goreleaser-ff69b4.svg)](https://goreleaser.com)
[![Code Quality](https://img.shields.io/badge/linter-golangci--lint-brightgreen.svg)](https://golangci-lint.run)
[![ZenTao Backend](https://img.shields.io/badge/zentao%20pms-v21.7%2B-orange.svg)](https://www.zentao.net)

[核心价值](#-为什么需要-zentao-cli) •
[功能特性](#-核心设计与价值主张) •
[安装指南](#-安装指南) •
[快速开始](#-快速开始) •
[命令手册](#-全量命令使用手册) •
[AI Agent 集成](#-ai-agent--大模型-tool-calling-集成) •
[Go SDK 使用](#-作为-go-sdk-独立调用-pkgzentao) •
[质量与测试](#-代码质量与自动化测试)

</div>

---

## 💡 为什么需要 zentao-cli？

在 AI 编码助手（如 Claude Code, ZCode, Cursor, OpenHands 等）与企业项目管理系统（ZenTao PMS）深度结合的日常开发场景中，传统的交互方式面临诸多痛点：

- ❌ **Web 会话频繁超时**：禅道基于 PHP 原生 Session（默认 24 分钟不活跃即回收），导致自动化脚本与 Agent 任务频繁被 `登录已超时，请重新登入!` 中断。
- ❌ **旧版 SDK 能力缺失且年久失修**：官方早期 PHP SDK 仅提供 20 余个简易 CRUD 接口，缺失**个人工作台、待办看板、活动流、多维过滤**等高频生产力能力。
- ❌ **非标准响应干扰大模型**：禅道原生错误多为 HTML 堆栈源码或双重转义字符串，直接破坏 LLM Function Calling 的结构化解析。

`zentao-cli`（可执行制品名为 `zentao`）基于禅道官方最新规范与底层接口彻底重构，不仅提供符合人体工程学的 CLI，更是一套成熟稳定的 **Go 语言原生 ZenTao SDK (`pkg/zentao`)**。

---

## 🌟 核心设计与价值主张

### 1. 唯一登录认证与全自动无感续约（Zero-Auth Subcommands）
- **仅需登录一次**：账号密码仅在 `zentao auth login` 执行时使用，随后安全持久化在本地 Profile 中（文件权限严格限制为 `0600`）。
- **零凭据污染**：所有业务子命令（`my`, `task`, `bug`, `project`, `product`, `todo` 等）**无需也不允许传递账号密码**。
- **透明自动恢复**：当服务端 Session 过期时，底层 HTTP 引擎**全自动触发二次握手重新登录、无缝重放业务请求并同步更新本地缓存**，上层调用完全无感。

### 2. 独立导出的 Go SDK (`pkg/zentao`)
- 底层通信与业务逻辑完全解耦为独立的公共包 `pkg/zentao`，零 CLI 依赖，外部 Go 项目可直接 `import "github.com/windosx/zentao-cli/pkg/zentao"`。
- 精准对齐禅道路由器的**位置参数映射序列**（Positional Parameter Mapping），根治 SQLSTATE 排序字段错位顽疾。

### 3. 全量个人工作台看板 (`zentao my` / `zentao todo`)
- 深度支持个人专属视图：**我的待办任务、指派给我的 Bug、日程待办事项 (Todo)、我的需求、我参与的项目以及操作活动流水**。
- 完整的待办生命周期流转（新建、开始、完成、关闭）。

### 4. Agent 友好的结构化协议与多格式输出
- **`-o json`（默认）**：统一响应信封 `{"ok": true, "outcome": "success", "data": ...}`，零冗余日志。
- **`-o table`**：自动解构嵌套资源为清晰对齐的 ASCII 表格。
- **`-o text`**：单行精简文本摘要，适合行级阅读。
- **`-o yaml` / `-o raw-json`**：满足多种自动化场景。
- **确定性退出码**：`0` (成功), `1` (API/业务错误), `2` (未登录), `3` (参数校验错误)。

### 5. 运行时工具探测与一键技能分发
- **`zentao schema`**：零网络开销动态导出命令与参数的 JSON Schema（包含参数类型、必填项、副作用与风险等级），专供 LLM 进行 Tool Calling 发现。
- **`zentao skill setup`**：一键将行业标准的 `SKILL.md` 注册分发至 `~/.zcode/skills/`、`~/.agents/skills/`、`~/.claude/skills/` 等各大 Agent 环境。

---

## 📦 安装指南

> **提示**：本项目名为 **`zentao-cli`**，可执行制品名统一为 **`zentao`**（brew/winget/choco/deb 安装后直接使用 `zentao` 命令；`go install` 因 Go 工具链按模块名命名，安装为 `zentao-cli`，可加一条 alias）。

### 1. macOS / Linux (Homebrew)
```bash
# 安装（tap 仓库 homebrew-tap 自动映射为短名 windosx/tap）
brew install windosx/tap/zentao-cli

# 安装后直接使用：zentao
```

### 2. Windows (WinGet / Chocolatey)
```powershell
# 方式 A：WinGet (Windows 10/11 官方推荐)
winget install windosx.zentao-cli

# 方式 B：Chocolatey
choco install zentao-cli

# 安装后直接使用：zentao
```

### 3. Linux (Debian / Ubuntu / APT)
```bash
# 下载对应架构的 deb 安装包并使用 apt 安装 (amd64 / arm64)
sudo apt-get install ./zentao-cli_*_linux_amd64.deb
# 安装后直接使用：zentao
```

### 4. 通过 Go Install 安装
```bash
# 安装（Go 工具链按模块名生成二进制，安装为 zentao-cli）
go install github.com/windosx/zentao-cli@latest

# 建议添加 alias 以统一使用 zentao 命令
# 在 ~/.zshrc 或 ~/.bashrc 中加入：
alias zentao=zentao-cli
```
> `go install` 构建的版本信息会从模块元数据自动读取（如 `v1.0.5`），与 GitHub Releases 中 goreleaser 注入的版本保持一致。

### 5. 下载预编译二进制 (GitHub Releases)
前往 [GitHub Releases](https://github.com/windosx/zentao-cli/releases) 下载适用于你操作系统的 tar.gz / zip 压缩包，解压后将 `zentao` 移动至系统 `PATH` 目录即可。

---

## 🚀 快速开始

### 1. 登录并绑定环境（仅需一次）

```bash
# 登录禅道服务器 (凭据将加密持久化于 ~/.config/zentao/profiles.json)
zentao auth login --url http://zentao.example.com --account testuser --password "Test@123456"
```

### 2. 日常高频操作（直接使用，零鉴权参数）

```bash
# 1. 查看指派给我的任务
zentao my task -o table

# 2. 查看指派给我的缺陷 (Bug)
zentao my bug -o table

# 3. 查看今天的日程待办
zentao my todo --type today -o table

# 4. 创建一条新待办并标记完成
zentao todo create --name "完成代码 Review" --pri 1
zentao todo finish --id 3

# 5. 查看项目下的进行中任务
zentao task list --project 109 --status doing -o table

# 6. 一键安装 Agent 技能到 AI 环境
zentao skill setup
```

---

## 📖 全量命令使用手册

全局参数（适用于所有命令）：

```bash
Flags:
  -u, --url string           禅道服务器 Base URL 地址 (默认从 profiles.json 读取)
  -m, --access-mode string   接口路由模式: GET (默认) 或 PATH_INFO
  -o, --output string        输出格式: json (默认), raw-json, yaml, table, text
  -k, --insecure             允许非安全的 HTTPS 连接 (忽略 SSL 证书校验)
  -c, --config string        指定配置文件路径 (默认: ~/.config/zentao/config.yaml)
      --timeout string       HTTP 请求超时时长 (默认 "30s")
```

---

### 1. 个人工作台与待办 (`my` / `todo`)

```bash
# 我的任务
zentao my task -o table                                      # 指派给我的任务 (默认)
zentao my task --type finishedBy -o table                    # 由我完成的任务
zentao my task --type openedBy -o table                      # 由我创建的任务
zentao my task --type undone -o table                        # 我的未完成任务

# 我的缺陷 (Bug)
zentao my bug -o table                                       # 指派给我的缺陷 (默认)
zentao my bug --type openedBy -o table                       # 由我创建的缺陷
zentao my bug --type resolvedBy -o table                     # 由我解决的缺陷

# 我的待办 (Todo)
zentao my todo -o table                                      # 全部待办
zentao my todo --type today -o table                         # 今天的待办
zentao my todo --type thisWeek -o table                      # 本周待办
zentao my todo --type before -o table                        # 逾期未完待办

# 待办事项管理 (todo)
zentao todo list --status wait -o table                      # 查询未开始的待办
zentao todo create --name "梳理架构设计" --date "2026-08-27" --pri 2
zentao todo start --id 3                                     # 标记开始 (doing)
zentao todo finish --id 3                                    # 标记完成 (done)
zentao todo close --id 3                                     # 标记关闭 (closed)

# 我的需求与项目
zentao my story --type assignedTo -o table                   # 指派给我的需求
zentao my project --status doing -o table                    # 我参与的进行中项目
zentao my dynamic --type today -o text                       # 今天的操作动态流水
```

---

### 2. 项目与任务管理 (`project` / `task`)

```bash
# 项目管理
zentao project list --status doing -o table                  # 查询进行中的项目
zentao project list --status all -o table                    # 查询全部项目
zentao project add --name "Sprint 2" --code "s2" --begin "2026-09-01" --end "2026-09-15"

# 任务管理
zentao task list --project 109 --status doing -o table       # 查询执行 109 下进行中的任务
zentao task list --project 109 --status all -o table         # 查询全部任务
zentao task params --project 109                             # 获取创建任务所需的模块与指派人元数据
zentao task create --project 109 --name "实现登录API" --assigned-to "testuser" --estimate 4.0
zentao task finish --id 657 --real 2.0 --comment "已完成单元测试覆盖"
```

---

### 3. 缺陷管理 (`bug`)

```bash
# 查询产品 Bug
zentao bug list --product 8 --browse-type unclosed -o table  # 未关闭的缺陷
zentao bug list --product 8 --browse-type assigntome -o table # 指派给我的缺陷
zentao bug params --product 8                                # 获取提交 Bug 的版本与元数据
zentao bug create --product 8 --title "登录页崩溃" --severity 2 --assigned-to "testuser"
zentao bug resolve --id 2862 --resolution fixed --comment "已修复并在本地验证"
```

---

### 4. 用户与部门管理 (`user` / `dept`)

```bash
zentao user list -o table                                    # 查询公司成员列表
zentao dept list -o table                                    # 查询部门层级结构
```

---

### 5. 认证与多环境管理 (`auth`)

```bash
zentao auth login --url <url> --account <acc> --password <pwd> [--name prod]
zentao auth list                                             # 列出所有已保存的环境 Profile
zentao auth switch --name <profile_name>                     # 切换当前激活的环境
zentao auth status -o table                                  # 查看当前会话状态
zentao auth logout                                           # 注销并清除本地缓存
```

---

## 🤖 AI Agent / 大模型 Tool Calling 集成

### 1. 静态技能挂载 (`SKILL.md`)
执行以下命令即可自动分发 Agent Skill：
```bash
zentao skill setup
```
该命令会自动将符合 Agent 生态标准的 `skills/zentao/SKILL.md` 写入当前用户的 Agent 目录：
- `~/.zcode/skills/zentao/SKILL.md`（ZCode 环境）
- `~/.agents/skills/zentao/SKILL.md`（通用 Agent 环境）
- `~/.claude/skills/zentao/SKILL.md`（Claude Desktop / Anthropic 环境）

### 2. 动态参数 Schema 探索
大模型可在未阅读完整文档前，通过探测指令动态获取命令契约：
```bash
zentao schema my --compact
zentao schema task --compact
```

---

## 💻 作为 Go SDK 独立调用 (`pkg/zentao`)

`pkg/zentao` 不依赖任何 CLI 框架，提供了开箱即用的类型化客户端与 Session 自动重试机制：

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/windosx/zentao-cli/pkg/zentao"
)

func main() {
	// 初始化禅道客户端
	client := zentao.New(zentao.Config{
		URL:        "http://zentao.example.com",
		Account:    "testuser",
		Password:   "Test@123456",
		AccessMode: zentao.AccessModeGET,
	})

	ctx := context.Background()

	// 首次登录握手
	if err := client.Login(ctx); err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	// 1. 查询指派给我的任务 (Session 超时时客户端会自动无感重新登录重试)
	tasksData, err := client.MyTasks(ctx, zentao.Params{"type": {"assignedTo"}})
	if err != nil {
		log.Fatalf("获取我的任务失败: %v", err)
	}
	fmt.Printf("任务数据: %s\n", string(tasksData))

	// 2. 创建一条新待办
	_, err = client.TodoCreate(ctx, zentao.Params{
		"name": {"使用 Go SDK 创建的待办"},
		"date": {"2026-08-27"},
		"pri":  {"2"},
	})
	if err != nil {
		log.Fatalf("创建待办失败: %v", err)
	}
	fmt.Println("待办创建成功！")
}
```

---

## 🛠️ 代码质量与自动化测试

本项目遵循生产级开源软件标准，设立了完整的质量门禁与端到端集成测试：

```bash
# 1. 运行 golangci-lint 与单元竞态检测
make check

# 2. 格式化代码 (gofmt)
make fmt

# 3. 运行真机端到端闭环测试 (Zero-Pollution 保证无残留测试脏数据)
make test-integration

# 4. 生成全量代码覆盖率统计报告 (已生成 coverage.html)
make coverage

# 5. 安装 Git pre-commit 钩子 (提交前强制执行 make check)
make install-hooks
```

### 🔐 安全运行集成测试（凭据注入）

集成测试会真实连接禅道服务器并自动清理测试数据。**凭据绝不硬编码进代码或 README**，按以下优先级安全注入：

1. **环境变量**（CI / 临时场景推荐）：
   ```bash
   ZENTAO_TEST_URL=https://zentao.example.com \
   ZENTAO_TEST_ACCOUNT=your-account \
   ZENTAO_TEST_PASSWORD=your-password \
   make test-integration
   ```
2. **本地 `.env` 文件**（开发便利）：
   ```bash
   cp .env.example .env   # 填入真实测试凭据
   make test-integration
   ```
   > `.env` 已被 `.gitignore` 排除，**绝不会进入 git 版本库**；仓库中仅保留 `.env.example` 占位模板。
3. **本地持久化 Profile**（最便捷）：
   - 已通过 `zentao auth login` 保存的 `~/.config/zentao/profiles.json`（位于用户目录，天然不进 git），测试自动复用其凭据。

> ⚠️ **安全红线**：任何真实账号、密码、内网地址**严禁**写入代码、README、测试文件或 `.env.example`。泄漏后应视为已泄露并立即轮换密码。

---

## 📄 开源许可证

本项目基于 [MIT License](LICENSE) 开源发布。
欢迎提交 Issue 与 Pull Request！
