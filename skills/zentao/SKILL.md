---
name: zentao
description: 禅道 (ZenTao PMS) 命令行工具。用于查看个人工作台与待办(my/todo)、创建/查询/解决Bug、管理任务、查看项目与产品、或查询状态。命令前缀：zentao。
metadata:
  cli_version: ">=1.0.0"
  category: pms
  requires:
    bins:
      - zentao
---

# 禅道 (ZenTao PMS) Agent Skill

使用此工具与企业禅道系统（ZenTao PMS）进行无缝交互与数据管理。

---

## 核心意图与命令映射

| 意图 | 推荐命令 | 关键参数 | 风险/副作用 |
|---|---|---|---|
| **查看指派给我的任务** | `zentao my task -o json` | `--type assignedTo` (默认) | read |
| **查看由我完成的任务** | `zentao my task --type finishedBy -o json` | `--type finishedBy` | read |
| **查看由我创建的任务** | `zentao my task --type openedBy -o json` | `--type openedBy` | read |
| **查看指派给我的Bug** | `zentao my bug -o json` | `--type assignedTo` (默认) | read |
| **查看由我创建的Bug** | `zentao my bug --type openedBy -o json` | `--type openedBy` | read |
| **查看我的日程待办** | `zentao my todo -o json` | `--type today\|thisWeek\|before\|all`, `--status` | read |
| **新建待办事项** | `zentao todo create --name <name> --date <YYYY-MM-DD> -o json` | `--name` (必填), `--date`, `--pri` | write |
| **开始待办事项** | `zentao todo start --id <id> -o json` | `--id` (必填) | write |
| **完成待办事项** | `zentao todo finish --id <id> -o json` | `--id` (必填) | write |
| **关闭待办事项** | `zentao todo close --id <id> -o json` | `--id` (必填) | write |
| **删除待办事项** | `zentao todo delete --id <id> -o json` | `--id` (必填) | destructive |
| **查看我的需求/故事** | `zentao my story -o json` | `--type assignedTo` (默认) | read |
| **查看我参与的项目** | `zentao my project -o json` | `--status doing\|all` | read |
| **查看我的活动流动态** | `zentao my dynamic -o text` | `--type today\|thisWeek` | read |
| **查看项目任务列表** | `zentao task list --project <id> -o json` | `--project` (必填), `--status` | read |
| **获取创建任务元数据** | `zentao task params --project <id> -o json` | `--project` (必填) | read |
| **创建新任务** | `zentao task create --project <id> --name <name> -o json` | `--project`, `--name`, `--assigned-to`, `--estimate` | write |
| **获取完成任务元数据** | `zentao task finish-params --id <id> -o json` | `--id` (必填) | read |
| **完成任务** | `zentao task finish --id <id> --real <hours> -o json` | `--id`, `--real`, `--comment` | write |
| **删除任务** | `zentao task delete --id <id> --project <id> -o json` | `--id` (必填) | destructive |
| **查看产品Bug列表** | `zentao bug list --product <id> -o json` | `--product` (必填), `--browse-type` | read |
| **获取创建Bug元数据** | `zentao bug params --product <id> -o json` | `--product` (必填) | read |
| **提交Bug** | `zentao bug create --product <id> --title <title> -o json` | `--product`, `--title`, `--severity`, `--assigned-to`, `--steps` | write |
| **获取解决Bug元数据** | `zentao bug resolve-params --id <id> -o json` | `--id` (必填) | read |
| **解决Bug** | `zentao bug resolve --id <id> --resolution fixed -o json` | `--id`, `--resolution`, `--comment` | write |
| **删除Bug** | `zentao bug delete --id <id> -o json` | `--id` (必填) | destructive |
| **查看产品列表** | `zentao product list -o json` | `--status noclosed\|all` | read |
| **获取创建产品元数据** | `zentao product params --program <id> -o json` | `--program` | read |
| **创建产品** | `zentao product add --name <name> --code <code> -o json` | `--name`, `--code`, `--po`, `--qd`, `--rd` | write |
| **查看项目列表** | `zentao project list -o json` | `--status doing\|all` | read |
| **获取创建项目元数据** | `zentao project params --program <id> -o json` | `--program` | read |
| **创建项目** | `zentao project add --name <name> --code <code> -o json` | `--name`, `--code`, `--begin`, `--end` | write |
| **查看公司成员列表** | `zentao user list -o json` | `--dept <id>` | read |
| **获取创建用户元数据** | `zentao user params --dept <id> -o json` | `--dept` | read |
| **创建用户** | `zentao user add --username <acc> --user-password <pwd> --realname <name> -o json` | `--username`, `--user-password`, `--realname` | write |
| **查看部门层级树** | `zentao dept list -o json` | `--parent <id>` | read |
| **添加部门** | `zentao dept add --parent <id> --name <name> -o json` | `--parent`, `--name` | write |
| **运行时探测工具Schema**| `zentao schema [subcommand] --compact -o json`| 探索工具参数定义与类型 | read |
| **检查登录状态** | `zentao auth status -o json` | 检查会话有效性 | read |
| **登录与凭证持久化** | `zentao auth login --url <url> --account <acc> --password <pwd>` | 登录并建立持久会话 | write |

---

## 最佳执行规范 (SOP)

1. **输出格式选择**：
   - 供 Agent 结构化消费与提取时，始终传递 **`-o json`**。
   - 供人类用户直接在终端查看时，可使用 **`-o table`**。
2. **免密调用与会话保持**：
   - 登录成功后，凭证与 Session ID 自动保存在 `~/.config/zentao/profiles.json` 中。
   - **严禁在调用业务子命令（如 `my task`、`bug create` 等）时重复传参 `--account` 或 `--password`**。
   - 若服务端的 PHP Session 超时（如空闲 24 分钟），底座通信引擎会自动透明重新登录续期，无需手动干预。
3. **参数预探测 (Schema/Params)**：
   - 在构造复杂的 `task create` 或 `bug create` 前，推荐先调用 `zentao task params --project <id>` 或 `zentao bug params --product <id>` 获取项目/产品下的真实模块 ID、指派人列表和构建版本。
4. **写操作回读确认**：
   - 任何写操作（`create`、`finish`、`resolve`、`close`）执行成功后，推荐调用对应模块的 `list` 命令进行一次回读验证，确保流转结果准确。
5. **分页控制**：
   - 所有列表查询命令默认返回前 100 条（`--limit 100`，`--page 1`）。
   - 如需拉取全量数据，传递 `--limit all`；如需遍历后续页，传递 `--page <页码>`。
