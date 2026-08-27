---
name: zentao
description: 禅道 (ZenTao PMS) 命令行工具。用于查看个人工作台与待办(my/todo)、缺陷全生命周期管理(bug)、任务全生命周期管理(task)、需求全生命周期管理(story)、项目与产品管理(project/product)、用户与部门管理(user/dept)、多环境认证(auth)。命令前缀：zentao。
when_to_use: 用户提到禅道、ZenTao、PMS系统，或要求查看/创建/修改/流转“指派给我的任务/Bug/需求”、“今日待办”、“提交/修改/解决Bug”、“创建/修改/指派/完成任务”、“创建/评审/变更需求”、“项目与产品管理”等敏捷项目管理工作时使用。
metadata:
  cli_version: ">=1.1.0"
  category: pms
  requires:
    bins:
      - zentao
---

# 禅道 (ZenTao PMS) Agent Skill

企业级禅道系统（ZenTao PMS）自动化交互与数据管理工具。

---

## 1. 核心执行哲学：CLI Help 与动态自省 (Self-Introspection)

> 💡 **单一事实源原则 (SSOT)**：
> CLI 内置的 `--help` 和 `schema` 是参数定义、类型约束与必填项的**唯一权威来源**。
> **严禁硬编码或猜测参数名**，在执行复杂操作前请充分利用 CLI 自省机制。

1. **查阅命令帮助**：执行 `zentao <module> [subcommand] --help` 查看最新的参数列表、简写与说明。
   ```bash
   zentao task create --help     # 查看创建任务支持的全部参数（如 --keywords, --mailto 等）
   zentao bug resolve --help     # 查看解决 Bug 支持的解决方案枚举
   zentao story review --help    # 查看评审需求的结果枚举
   ```
2. **探测工具元数据 Schema**：执行 `zentao schema [command] --compact -o json` 供 Agent 结构化探测子命令与参数。
3. **获取服务端动态字典**：创建或修改实体前，调用 `zentao <module> params` 获取服务端实时字典（模块树、指派人列表、影响版本等）：
   ```bash
   zentao task params --project <id> -o json    # 获取项目下的模块与指派人字典
   zentao bug params --product <id> -o json     # 获取产品下的分支、模块、版本字典
   zentao story params --product <id> -o json   # 获取需求分类、计划、来源字典
   ```

---

## 2. 模块与核心意图速查

| 业务实体 | 命令前缀 | 核心操作能力 | 常见子命令 |
|---|---|---|---|
| **个人工作台** | `zentao my` | 查看当前登录用户的各类待办事项与流水 | `task`, `bug`, `story`, `todo`, `project`, `dynamic` |
| **待办事项** | `zentao todo` | 个人日程待办全生命周期管理 | `list`, `view`, `create`, `edit`, `start`, `finish`, `close`, `activate`, `assign`, `delete` |
| **任务** | `zentao task` | 项目/执行下的开发与测试任务流转 | `list`, `view`, `params`, `create`, `edit`, `start`, `pause`, `restart`, `finish`, `close`, `cancel`, `activate`, `assign`, `delete` |
| **缺陷** | `zentao bug` | 产品缺陷提交、流转与闭环 | `list`, `view`, `params`, `create`, `edit`, `resolve`, `close`, `activate`, `confirm`, `assign`, `delete` |
| **需求** | `zentao story` | 敏捷需求/用户故事全生命周期管理 | `list`, `view`, `params`, `create`, `edit`, `review`, `change`, `close`, `activate`, `assign`, `delete` |
| **项目/执行** | `zentao project` | 敏捷迭代与瀑布项目管理 | `list`, `view`, `params`, `create`, `edit`, `start`, `suspend`, `activate`, `close`, `delete` |
| **产品** | `zentao product` | 产品线与产品定义管理 | `list`, `view`, `params`, `create`, `edit`, `close`, `activate`, `delete` |
| **用户与部门** | `zentao user` / `dept` | 组织架构、部门树与成员管理 | `user {list\|view\|params\|create\|edit\|delete}`, `dept {list\|create\|edit\|delete}` |
| **认证与环境** | `zentao auth` | 多环境 Profile 切换与凭据安全持久化 | `login`, `list`, `switch`, `status`, `logout` |

---

## 3. 黄金工作流模式 (Workflow Patterns)

### 模式 A：个人工作台巡检
```bash
zentao my todo --type today -o json          # 今日待办
zentao my task --type assignedTo -o json     # 指派给我的进行中/待办任务
zentao my bug --type assignedTo -o json      # 指派给我的未解决缺陷
zentao my story --type assignedTo -o json    # 指派给我的需求
```

### 模式 B：规范的实体流转闭环（以 Bug 为例）
1. **探测元数据**：`zentao bug params --product <id> -o json`（获取可用的 `openedBuild`、`module` 与 `assignedTo`）。
2. **提交实体**：`zentao bug create --product <id> --title <title> --keywords <keywords> -o json`（通过 `--help` 查看更多选填字段）。
3. **编辑/更新**：`zentao bug edit --id <id> --comment <修改说明> -o json`。
4. **解决/流转**：`zentao bug resolve --id <id> --resolution fixed -o json`。
5. **关闭闭环**：`zentao bug close --id <id> -o json`。

---

## 4. 全局执行规范

1. **结构化输出**：供 Agent 处理时，**始终附加 `-o json`**。
2. **会话与免密**：凭据保存于系统 Keyring，Session 自动透明续期，业务命令**严禁重复传递 `--account` / `--password`**。
3. **分页控制**：默认返回 100 条。拉取全量传递 `--limit all`；分页传递 `--page <页码> --limit <条数>`。
4. **遇到参数疑问**：遇到不确定的参数名或枚举值时，第一步执行 `zentao <command> [subcommand] --help`，按帮助输出调整传参。
