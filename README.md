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

## 💡 为什么创建这个项目？（背景与选型指南）

### 项目背景与核心痛点

在 AI 编码助手（如 Claude Code, ZCode, Cursor, OpenHands 等）与企业项目管理系统（ZenTao PMS）深度集成的日常开发与自动化运维中，开发者与 Agent 经常面临以下实际挑战：

1. **历史与存量版本无 RESTful API 2.0 支持**：
   禅道官方自 **ZenTao 21.7.6**（2025 年 10 月）及 **22.x** 起才开始引入 RESTful API v2 (`/api.php/v2`)。国内大量企业和团队的生产私有化部署仍停留在 **ZenTao ≤ 21.7.5（包括 21.7.0, 21.6, 20.x, 19.x, 18.x...）**，这些版本完全不存在 API v2 模块。
2. **Web 会话频繁超时中断 AI 工作流**：
   禅道基于 PHP 原生 Session（默认 24 分钟无操作即回收），导致长时间运行的 Agent 任务频繁被 `登录已超时，请重新登入!` 中断并失败。
3. **个人工作台与待办日历能力缺失**：
   API v2 主要是对核心实体（产品/项目/Bug/任务）的基础 CRUD 门面路由，缺失了**个人工作台（`my/*` 指派给我的/我创建的/今日看板）**、**日程待办看板（`todo/*`）**、**操作动态流水（`dynamic`）**等开发者与 Agent 每天高频使用的生产力端点。
4. **单二进制与安全钥匙串需求**：
   现代开发与 CI/CD 流程渴望零运行环境依赖（无需 Node.js / Bun / npm 运行时）的单一跨平台可执行文件，以及系统原生钥匙串（macOS Keychain / Windows Credential Manager / Linux Secret Service）级别的凭据存储安全。

`zentao-cli`（制品名为 `zentao`）直接对齐禅道底层核心控制器（Native Web Controller JSON 通道）与官方 PHP SDK，提供全生命周期的 Go SDK (`pkg/zentao`) 与全平台 CLI。

---

### 选型建议：我应该用哪个 CLI？

- 🌟 **建议使用官方 [easysoft/zentao-cli](https://github.com/easysoft/zentao-cli) 的场景**：
  - 你的团队正在运行 **最新版 ZenTao ≥ 21.7.6 / 22.x**，且服务端已开启并配置好 RESTful API v2 模块；
  - 本地已具备 Node.js / Bun 运行时，主要通过 `npx` 或 npm 生态进行交互；
  - 仅需要标准实体（产品、项目、需求、Bug、任务）的基础 CRUD 管理。

- ⚡ **推荐使用本项目 [windosx/zentao-cli](https://github.com/windosx/zentao-cli) 的场景**：
  - 你的团队运行在 **ZenTao ≤ 21.7.5（如 18.x, 20.x, 21.0~21.7 等广泛存量部署）** 或未开放 API v2 的私有化实例上；
  - 面向 **AI Agent（Claude Code, Cursor, ZCode 等）长时间自主运行**，需要**零凭据泄漏与会话超时无感自动重登续约（Zero-Auth）**；
  - 深度依赖**个人工作台看板（`my` 指派给我/我创建的）**、**今日待办日历（`todo`）**与**操作活动流（`dynamic`）**；
  - 追求**单静态二进制开箱即用**（支持 Homebrew Cask, WinGet, Chocolatey 一键安装），无需 Node.js/Bun 运行时；
  - 需要在 Go 后端微服务或自动化插件中直接引用纯 Go SDK (`pkg/zentao`)。

---

## 🌟 核心设计与价值主张

### 1. 唯一登录认证与全自动无感续约（Zero-Auth Subcommands）
- **仅需登录一次**：账号密码仅在 `zentao auth login` 执行时使用，密码安全托管于系统原生钥匙串（macOS Keychain / Windows Credential Manager / Linux Secret Service；无桌面钥匙串服务时安全回退至权限受限的本地文件），会话凭据保存在本地 Profile 中。
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

> **提示**：本项目名为 **`zentao-cli`**，可执行制品名统一为 **`zentao`**（一键脚本/brew/winget/choco 安装后直接使用 `zentao` 命令；`go install` 因 Go 工具链按模块名命名，安装为 `zentao-cli`，可加一条 alias）。

### 1. 一键脚本安装（Linux & macOS 推荐，自动识别架构）
```bash
curl -fsSL https://raw.githubusercontent.com/windosx/zentao-cli/main/install.sh | bash

# 安装后直接使用：zentao
```

### 2. macOS / Linux (Homebrew)
```bash
# 安装（tap 仓库 homebrew-tap 自动映射为短名 windosx/tap）
brew install windosx/tap/zentao-cli

# 安装后直接使用：zentao
```

### 3. Windows (WinGet / Chocolatey / 手动下载)
```powershell
# 方式 A：WinGet (Windows 10/11 官方推荐)
# 注意：包已提交至 winget-pkgs 官方仓库，首次发布需通过官方审核合并后生效
winget install windosx.zentao-cli

# 方式 B：Chocolatey（即将上架）
choco install zentao-cli

# 方式 C：手动下载（立即可用）
# 从 GitHub Releases 下载 zentao-cli-<版本>-windows-<架构>.zip，
# 解压后将 zentao.exe 所在目录加入 PATH 即可使用 zentao 命令

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
> `go install` 构建的版本信息会从模块元数据自动读取（如 `v1.0.8`），与 GitHub Releases 中 goreleaser 注入的版本保持一致。

### 5. 下载预编译二进制 (GitHub Releases)
前往 [GitHub Releases](https://github.com/windosx/zentao-cli/releases) 下载适用于你操作系统的 tar.gz / zip 压缩包，解压后将 `zentao` 移动至系统 `PATH` 目录即可。

---

## 🚀 快速开始

### 1. 登录并绑定环境（仅需一次）

```bash
# 登录禅道服务器 (密码存入系统钥匙串，会话持久化于 ~/.config/zentao/profiles.json)
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

列表查询通用分页参数（适用于所有 list 系列命令）：
      --page int             页码，从 1 开始 (默认 1)
      --limit string         每页返回条数，默认 100；传 all 或 0 表示拉取全量 (默认 "100")
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
zentao todo delete --id 3                                    # 删除待办事项

# 待办管理
zentao todo list --type today -o table                       # 今日待办
zentao todo view --id 12 -o json                             # 查看待办详情
zentao todo create --name "代码评审" --date "2026-08-27"      # 创建待办
zentao todo edit --id 12 --name "代码评审与验证"              # 修改待办
zentao todo start --id 12                                    # 开始待办
zentao todo finish --id 12                                   # 完成待办
zentao todo close --id 12                                    # 关闭待办

# 我的需求与项目
zentao my story --type assignedTo -o table                   # 指派给我的需求
zentao my project --status doing -o table                    # 我参与的进行中项目
zentao my dynamic --type today -o text                       # 今天的操作动态流水
```

---

### 2. 项目、产品、需求与任务管理 (`project` / `product` / `story` / `task`)

```bash
# 产品管理
zentao product list --status noclosed -o table               # 查询正常运营中的产品
zentao product view --id 8 -o json                           # 查看产品详情
zentao product params --program 0                            # 获取创建产品所需的元数据字典
zentao product create --name "移动应用" --code "app" --po "po_user" # 创建产品
zentao product edit --id 8 --name "移动应用 Pro"             # 修改产品
zentao product close --id 8                                  # 关闭产品

# 项目管理
zentao project list --status doing -o table                  # 查询进行中的项目
zentao project view --id 109 -o json                         # 查看项目详情
zentao project params --program 0                            # 获取创建项目所需的元数据字典
zentao project create --name "Sprint 2" --code "s2" --begin "2026-09-01" --end "2026-09-15"
zentao project edit --id 109 --name "Sprint 2 (已调整)"      # 修改项目
zentao project start --id 109                                # 开始项目
zentao project suspend --id 109                              # 挂起项目
zentao project activate --id 109                             # 激活项目
zentao project close --id 109                                # 关闭项目

# 需求管理
zentao story list --product 8 -o table                       # 查询产品需求列表
zentao story view --id 55 -o json                            # 查看需求详情
zentao story params --product 8                              # 获取创建需求元数据
zentao story create --product 8 --title "支持微信支付" --keywords "支付,微信" --spec "用户可在结算页使用微信支付"
zentao story edit --id 55 --keywords "支付,微信,扫码"        # 修改需求/更新关键词
zentao story review --id 55 --result pass                    # 评审需求
zentao story change --id 55 --spec "更新后的需求规格说明"      # 变更需求
zentao story close --id 55 --reason done                     # 关闭需求

# 任务管理
zentao task list --project 109 --status doing -o table       # 查询执行 109 下进行中的任务
zentao task view --id 657 -o json                            # 查看任务详情
zentao task params --project 109                             # 获取创建任务所需的模块与指派人元数据
zentao task create --project 109 --name "实现登录API" --keywords "认证,JWT" --assigned-to "testuser" --estimate 4.0
zentao task edit --id 657 --keywords "认证,JWT,OAuth"        # 修改任务/更新关键词
zentao task start --id 657 --real-started "2026-08-27 09:30:00" # 开始执行任务
zentao task pause --id 657                                   # 暂停任务
zentao task restart --id 657                                 # 重启/继续任务
zentao task finish-params --id 657                           # 获取完成任务所需的当前状态与表单元数据
zentao task finish --id 657 --real 2.0 --comment "已完成单元测试覆盖"
zentao task close --id 657                                   # 关闭任务
zentao task cancel --id 657 --comment "任务取消说明"          # 取消任务
zentao task delete --id 657 --project 109                    # 删除指定任务
```

---

### 3. 缺陷管理 (`bug`)

```bash
# 查询产品 Bug
zentao bug list --product 8 --browse-type unclosed -o table  # 未关闭的缺陷
zentao bug view --id 2862 -o json                            # 查看 Bug 详情
zentao bug list --product 8 --browse-type assigntome -o table # 指派给我的缺陷
zentao bug params --product 8                                # 获取提交 Bug 的版本与元数据
zentao bug create --product 8 --title "登录页崩溃" --keywords "崩溃,前端" --severity 2 --assigned-to "testuser"
zentao bug edit --id 2862 --keywords "崩溃,前端,Safari"       # 修改 Bug/更新关键词
zentao bug resolve-params --id 2862                          # 获取解决 Bug 的方案与构建版本元数据
zentao bug resolve --id 2862 --resolution fixed --comment "已修复并在本地验证"
zentao bug close --id 2862                                   # 关闭 Bug
zentao bug activate --id 2862 --comment "在iOS端重现，重新打开" # 激活 Bug
zentao bug delete --id 2862                                  # 删除指定 Bug
```

---

### 4. 用户与部门管理 (`user` / `dept`)

```bash
zentao user list -o table                                    # 查询公司成员列表
zentao user view --id 12 -o json                             # 查看用户详情
zentao user params --dept 1                                  # 获取创建用户所需的部门树与角色元数据
zentao user create --username "tom" --user-password "pwd" --realname "Tom"
zentao user edit --id 12 --realname "Thomas"                 # 修改用户信息
zentao user restore --id 12                                  # 恢复已删除的用户
zentao dept list -o table                                    # 查询部门层级结构
zentao dept create --parent 1 --name "前端组"                # 添加子部门
zentao dept edit --id 3 --name "移动端研发组"                # 修改部门
```

---

### 5. 系统回收站管理 (`trash`)

```bash
zentao trash list -o table                                   # 查询回收站中全部已删除对象
zentao trash list --type task -o table                       # 仅查询已删除的任务
zentao trash restore --action-id 36423                       # 按删除动作 ID 恢复对象
zentao trash hide-one --action-id 36423                      # 在回收站中隐藏指定的删除记录
zentao trash hide-all                                        # 清空/隐藏回收站中的全部删除记录
```

> 💡 **提示**：除了 `zentao trash` 全局命令外，各业务模块均支持快捷恢复：`zentao task/bug/story/project/product/user/todo restore --id <ID>`。

---

### 6. 认证与多环境管理 (`auth`)

```bash
zentao auth login --url <url> --account <acc> --password <pwd> [--name prod]
zentao auth list                                             # 列出所有已保存的环境 Profile
zentao auth switch --name <profile_name>                     # 切换当前激活的环境
zentao auth status -o table                                  # 查看当前会话状态
zentao auth logout                                           # 注销并清除本地缓存
```

---

### 7. 配置文件管理 (`config`)

```bash
zentao config init --path .zentao.yaml                       # 初始化当前目录配置文件模板
zentao config show -o table                                  # 查看当前生效配置参数与会话信息
```

---

### 8. 版本与兼容性查看 (`version`)

本项目遵循 **SemVer 2.0 构建元数据规范 (Build Metadata)**，完整版本格式为 `vX.Y.Z+<zentao_version>`（例如 `v1.0.8+21.7`）：
- `vX.Y.Z`：CLI 与 Go SDK 自身的语义化迭代版本；
- `+21.7`：所深度对齐与兼容的禅道官方底层 API / PHP SDK 版本。

#### 服务端版本兼容矩阵

| 禅道服务端版本 | 兼容状态 | 说明 |
|---|---|---|
| **ZenTao 21.x**（如 `21.7+`） | 🟢 **完全支持 / 已深度验证** | 生产验证版本，全量功能可用（个人工作台、待办看板、活动流、任务/Bug/项目/产品 CRUD） |
| **ZenTao 20.x** | 🟡 **Best-Effort** | 核心实体与工作台可用，部分较新字段可能缺失 |
| **ZenTao ≤ 19.x** | ⚪ **未做回归测试** | 基础 CRUD 走早期路由兼容，可能受部分表结构调整影响 |

#### 官方 API 通道评估结论

- **关于官方 RESTful API v1 (`/api.php/v1`)**：早期接口，为功能极简子集且已被官方废弃，缺少 `my/*`（个人工作台任务/Bug/待办看板）与 `dynamic`（活动流水）等核心生产力端点；且创建/编辑任务等操作无法透传 `keywords` 等表单字段。
- **关于官方 RESTful API v2 (`/api.php/v2`)**：禅道官方自 **21.7.6+**（2025 年 10 月）及 **22.x** 起引入，本质上是底层 Web 控制器方法的 RESTful 门面路由。由于早期存量版本（≤ 21.7.5）完全未包含 v2 模块，且 v2 的 Token 同样受服务端 Session 生命周期限制，本项目统一采用官方 PHP SDK 所使用的原生 Web 控制器 JSON 通道，保证最大范围的跨版本全兼容与全量工作台功能覆盖。

```bash
zentao version -o table
```

输出示例：
```text
KEY             VALUE
---             -----
arch            arm64
buildDate       2026-08-27T08:00:00Z
fullVersion     1.0.8+21.7
gitCommit       18a46d1
goVersion       go1.25.3
os              darwin
sdkVersion      zentaopms_21.7_20250516
version         1.0.8
zentaoCompat    v21.7+
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

### 🔒 安全设计与注意事项

1. **凭证存储**：
   - 生产环境中，`zentao auth login` 会将密码委托给系统原生钥匙串（macOS Keychain、Windows Credential Manager 或 Linux Secret Service）。
   - 在无图形钥匙串服务的无头 Linux / 容器环境中，会自动降级为权限受限的本地文件存储（`0600` 文件权限）；亦可设置环境变量 `ZENTAO_NO_KEYRING=1` 显式控制。
2. **删除操作与 HTTP 语义**：
   - `task delete`、`bug delete`、`todo delete` 遵循禅道原生路由设计（带 `confirm=yes` 的控制路由），在某些反向代理或具有主动预取策略的中间网关环境下，请确保该类路由不被预加载。

---

## 📄 开源许可证

本项目采用双许可结构：

- **仓库整体**（`cmd/`、`internal/`、构建与发布配置等）：基于 [MIT License](LICENSE) 开源发布。
- **`pkg/zentao`（Go 语言 ZenTao 客户端）**：移植自禅道官方 PHP SDK（`zentaopms_21.7_20250516` 的 `sdk/php/zentao.php`），作为衍生作品遵循 [Z Public License (ZPL) 1.2](LICENSE-ZPL) 分发，并保留上游版权署名（详见 [NOTICE](NOTICE)）。

> 说明：ZPL 1.2 允许商业与非商业使用、修改及再分发；基于禅道 HTTP API 独立开发的应用本可按 ZPL 第 7 条使用自有许可证，但 `pkg/zentao` 包含移植自官方 SDK 的代码逻辑，故谨慎地采用 ZPL 1.2。

欢迎提交 Issue 与 Pull Request！
