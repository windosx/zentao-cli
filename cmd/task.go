package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "管理禅道任务",
	Long:  "查询项目/执行下的任务列表、查看任务详情、获取元数据、创建、编辑或执行生命周期操作（开始、暂停、重启、完成、关闭、取消、激活、指派、删除）。",
}

var (
	taskProjectID    string
	taskExecutionID  string
	taskParentID     string
	taskID           string
	taskStatus       string
	taskOrderBy      string
	taskName         string
	taskType         string
	taskPri          string
	taskEstimate     string
	taskLeft         string
	taskConsumed     string
	taskAssignedTo   string
	taskDesc         string
	taskModuleID     string
	taskStoryID      string
	taskKeywords     string
	taskMailto       string
	taskDeadline     string
	taskEstStarted   string
	taskRealStarted  string
	taskReal         string
	taskComment      string
	taskFinishedDate string
)

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询指定项目/执行的任务列表（支持按状态与指派人筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskProjectID == "" {
			return fmt.Errorf("--project 是必填参数（传入项目 ID 或执行/迭代 ID）")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"executionID": {taskProjectID},
			"projectID":   {taskProjectID},
			"project":     {taskProjectID},
		}
		if taskStatus != "" {
			params.Set("status", taskStatus)
		}
		if taskOrderBy != "" {
			params.Set("orderBy", taskOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.TaskList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定任务的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TaskView(ctx, taskID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取指定项目下创建任务所需的元数据与字典（模块、指派人列表等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskProjectID == "" {
			return fmt.Errorf("--project 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TaskCreateParams(ctx, taskProjectID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "在指定项目/执行下创建新任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskProjectID == "" {
			return fmt.Errorf("--project 是必填参数")
		}
		if taskName == "" {
			return fmt.Errorf("--name 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"project": {taskProjectID},
			"name":    {taskName},
		}
		if taskExecutionID != "" {
			params.Set("execution", taskExecutionID)
		}
		if taskParentID != "" && taskParentID != "0" {
			params.Set("parent", taskParentID)
		}
		if taskType != "" {
			params.Set("type", taskType)
		}
		if taskPri != "" {
			params.Set("pri", taskPri)
		}
		if taskEstimate != "" {
			params.Set("estimate", taskEstimate)
		}
		leftVal := taskLeft
		if leftVal == "" && taskEstimate != "" {
			leftVal = taskEstimate
		}
		if leftVal != "" {
			params.Set("left", leftVal)
		}
		if taskAssignedTo != "" {
			params.Set("assignedTo", taskAssignedTo)
		}
		if taskDesc != "" {
			params.Set("desc", taskDesc)
		}
		if taskModuleID != "" {
			params.Set("module", taskModuleID)
		} else {
			params.Set("module", "0")
		}
		if taskStoryID != "" {
			params.Set("story", taskStoryID)
		} else {
			params.Set("story", "0")
		}
		if taskKeywords != "" {
			params.Set("keywords", taskKeywords)
		}
		if taskMailto != "" {
			params.Set("mailto", taskMailto)
		}
		if taskDeadline != "" {
			params.Set("deadline", taskDeadline)
		}
		if taskEstStarted != "" {
			params.Set("estStarted", taskEstStarted)
		}

		data, err := client.TaskCreate(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定任务的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		// ZenTao's task/edit action requires full form validation on submit
		// (including execution, type, name, status, assignedTo, estimate, left, consumed).
		// We fetch existing task details to automatically preserve non-updated required fields.
		existingData, err := client.TaskView(ctx, taskID)
		if err == nil {
			var viewResp struct {
				Task struct {
					Name       string `json:"name"`
					Type       string `json:"type"`
					Pri        any    `json:"pri"`
					Estimate   any    `json:"estimate"`
					Left       any    `json:"left"`
					Consumed   any    `json:"consumed"`
					AssignedTo string `json:"assignedTo"`
					Desc       string `json:"desc"`
					Module     any    `json:"module"`
					Story      any    `json:"story"`
					Keywords   string `json:"keywords"`
					Mailto     string `json:"mailto"`
					Deadline   any    `json:"deadline"`
					EstStarted any    `json:"estStarted"`
					Status     string `json:"status"`
					Execution  any    `json:"execution"`
					Parent     any    `json:"parent"`
				} `json:"task"`
			}
			if err := json.Unmarshal(existingData, &viewResp); err == nil && viewResp.Task.Name != "" {
				t := viewResp.Task
				valToStr := func(v any) string {
					if v == nil {
						return ""
					}
					s := fmt.Sprint(v)
					if s == "<nil>" || s == "null" {
						return ""
					}
					return s
				}

				if !cmd.Flags().Changed("execution") && !cmd.Flags().Changed("project") {
					taskExecutionID = valToStr(t.Execution)
				}
				if !cmd.Flags().Changed("parent") {
					taskParentID = valToStr(t.Parent)
				}
				if !cmd.Flags().Changed("name") {
					taskName = t.Name
				}
				if !cmd.Flags().Changed("type") {
					taskType = t.Type
				}
				if !cmd.Flags().Changed("pri") {
					taskPri = valToStr(t.Pri)
				}
				if !cmd.Flags().Changed("estimate") {
					taskEstimate = valToStr(t.Estimate)
				}
				if !cmd.Flags().Changed("left") {
					taskLeft = valToStr(t.Left)
				}
				if !cmd.Flags().Changed("consumed") {
					taskConsumed = valToStr(t.Consumed)
				}
				if !cmd.Flags().Changed("assigned-to") {
					taskAssignedTo = t.AssignedTo
				}
				if !cmd.Flags().Changed("desc") {
					taskDesc = t.Desc
				}
				if !cmd.Flags().Changed("module") {
					taskModuleID = valToStr(t.Module)
				}
				if !cmd.Flags().Changed("story") {
					taskStoryID = valToStr(t.Story)
				}
				if !cmd.Flags().Changed("keywords") {
					taskKeywords = t.Keywords
				}
				if !cmd.Flags().Changed("mailto") {
					taskMailto = t.Mailto
				}
				if !cmd.Flags().Changed("deadline") {
					taskDeadline = valToStr(t.Deadline)
				}
				if !cmd.Flags().Changed("est-started") {
					taskEstStarted = valToStr(t.EstStarted)
				}
				if !cmd.Flags().Changed("status") {
					taskStatus = t.Status
				}
			}
		}

		params := zentao.Params{}
		exec := taskExecutionID
		if exec == "" {
			exec = taskProjectID
		}
		if exec != "" {
			params.Set("execution", exec)
			params.Set("project", exec)
		}
		if taskParentID != "" {
			params.Set("parent", taskParentID)
		}
		if taskName != "" {
			params.Set("name", taskName)
		}
		if taskType != "" {
			params.Set("type", taskType)
		}
		if taskPri != "" {
			params.Set("pri", taskPri)
		}
		if taskEstimate != "" {
			params.Set("estimate", taskEstimate)
		}
		if taskLeft != "" {
			params.Set("left", taskLeft)
		}
		if taskConsumed != "" {
			params.Set("consumed", taskConsumed)
		}
		if taskAssignedTo != "" {
			params.Set("assignedTo", taskAssignedTo)
		}
		if taskDesc != "" {
			params.Set("desc", taskDesc)
		}
		if taskModuleID != "" {
			params.Set("module", taskModuleID)
		}
		if taskStoryID != "" {
			params.Set("story", taskStoryID)
		}
		if taskKeywords != "" {
			params.Set("keywords", taskKeywords)
		}
		if taskMailto != "" {
			params.Set("mailto", taskMailto)
		}
		if taskDeadline != "" {
			params.Set("deadline", taskDeadline)
		}
		if taskEstStarted != "" {
			params.Set("estStarted", taskEstStarted)
		}
		if taskStatus != "" {
			params.Set("status", taskStatus)
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskEdit(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskStartCmd = &cobra.Command{
	Use:   "start",
	Short: "开始执行任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskRealStarted != "" {
			params.Set("realStarted", taskRealStarted)
		} else {
			params.Set("realStarted", time.Now().Format("2006-01-02 15:04:05"))
		}
		if taskLeft != "" {
			params.Set("left", taskLeft)
		}
		if taskConsumed != "" {
			params.Set("consumed", taskConsumed)
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskStart(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "暂停正在执行的任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskPause(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "继续/重启已暂停的任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskLeft != "" {
			params.Set("left", taskLeft)
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskRestart(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskFinishParamsCmd = &cobra.Command{
	Use:   "finish-params",
	Short: "获取完成任务所需的元数据字典与当前任务状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TaskFinishParams(ctx, taskID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "完成任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		realConsumed := taskReal
		if realConsumed == "" {
			realConsumed = "1.0"
		}

		now := time.Now()
		finishedDate := taskFinishedDate
		if finishedDate == "" {
			finishedDate = now.Format("2006-01-02 15:04:05")
		} else if !strings.Contains(finishedDate, " ") {
			finishedDate = finishedDate + " " + now.Format("15:04:05")
		}

		params := zentao.Params{
			"real":            {realConsumed},
			"currentConsumed": {realConsumed},
			"consumed":        {realConsumed},
			"finishedDate":    {finishedDate},
			"realStarted":     {now.Format("2006-01-02 09:00:00")},
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskFinish(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭已完成或已取消的任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskClose(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "取消任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskCancel(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已关闭或已取消的任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if taskLeft != "" {
			params.Set("left", taskLeft)
		}
		if taskAssignedTo != "" {
			params.Set("assignedTo", taskAssignedTo)
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskActivate(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "指派任务给指定用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if taskAssignedTo == "" {
			return fmt.Errorf("--assigned-to 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"assignedTo": {taskAssignedTo},
		}
		if taskLeft != "" {
			params.Set("left", taskLeft)
		}
		if taskConsumed != "" {
			params.Set("consumed", taskConsumed)
		}
		if taskComment != "" {
			params.Set("comment", taskComment)
		}

		data, err := client.TaskAssign(ctx, taskID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		projectID := taskProjectID
		if projectID == "" {
			projectID = "0"
		}

		data, err := client.TaskDelete(ctx, projectID, taskID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var taskRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "task", taskID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	taskListCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 ID 或 执行/迭代 ID (必填)")
	taskListCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号筛选")
	taskListCmd.Flags().StringVar(&taskStatus, "status", "all", "任务状态过滤: all (全部), undone/unclosed (未完成), wait (未开始), doing (进行中), done (已完成), pause (已暂停), cancel (已取消), closed (已关闭), needconfirm (需求变动待确认)")
	taskListCmd.Flags().StringVar(&taskOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, id_asc, pri_asc, deadline_asc, estimate_desc, consumed_desc)")
	addPaginationFlags(taskListCmd)

	taskViewCmd.Flags().StringVar(&taskID, "id", "", "要查看的任务 ID (必填)")

	taskParamsCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 / 执行 ID (必填)")

	taskCreateCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 / 执行 ID (必填)")
	taskCreateCmd.Flags().StringVar(&taskExecutionID, "execution", "", "所属执行/迭代 ID (若与 --project 相同可省略)")
	taskCreateCmd.Flags().StringVar(&taskParentID, "parent", "0", "所属父任务 ID")
	taskCreateCmd.Flags().StringVar(&taskName, "name", "", "任务名称 (必填)")
	taskCreateCmd.Flags().StringVar(&taskType, "type", "devel", "任务类型 (design: 设计, devel: 开发, test: 测试, study: 研究, discuss: 讨论, ui: 界面, affair: 事务, misc: 其他)")
	taskCreateCmd.Flags().StringVar(&taskPri, "pri", "3", "任务优先级 (1=最高, 2=高, 3=中, 4=低)")
	taskCreateCmd.Flags().StringVar(&taskEstimate, "estimate", "", "预计工时 (例如: 4.5)")
	taskCreateCmd.Flags().StringVar(&taskLeft, "left", "", "预计剩余工时 (若不提供，默认自动同步为 estimate 的值)")
	taskCreateCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号")
	taskCreateCmd.Flags().StringVar(&taskDesc, "desc", "", "任务详细描述说明")
	taskCreateCmd.Flags().StringVar(&taskModuleID, "module", "0", "所属模块 ID")
	taskCreateCmd.Flags().StringVar(&taskStoryID, "story", "0", "关联所属需求/故事 ID")
	taskCreateCmd.Flags().StringVar(&taskKeywords, "keywords", "", "任务关键词 (例如: 登录, 权限)")
	taskCreateCmd.Flags().StringVar(&taskMailto, "mailto", "", "抄送给的用户账号列表 (逗号分隔)")
	taskCreateCmd.Flags().StringVar(&taskDeadline, "deadline", "", "截止日期 (格式: YYYY-MM-DD)")
	taskCreateCmd.Flags().StringVar(&taskEstStarted, "est-started", "", "预计开始日期 (格式: YYYY-MM-DD)")

	taskEditCmd.Flags().StringVar(&taskID, "id", "", "要修改的任务 ID (必填)")
	taskEditCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 ID")
	taskEditCmd.Flags().StringVar(&taskExecutionID, "execution", "", "所属执行/迭代 ID")
	taskEditCmd.Flags().StringVar(&taskParentID, "parent", "", "所属父任务 ID")
	taskEditCmd.Flags().StringVar(&taskName, "name", "", "任务名称")
	taskEditCmd.Flags().StringVar(&taskType, "type", "", "任务类型")
	taskEditCmd.Flags().StringVar(&taskPri, "pri", "", "任务优先级 (1=最高, 2=高, 3=中, 4=低)")
	taskEditCmd.Flags().StringVar(&taskEstimate, "estimate", "", "预计工时")
	taskEditCmd.Flags().StringVar(&taskLeft, "left", "", "剩余工时")
	taskEditCmd.Flags().StringVar(&taskConsumed, "consumed", "", "已消耗工时")
	taskEditCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号")
	taskEditCmd.Flags().StringVar(&taskDesc, "desc", "", "任务描述")
	taskEditCmd.Flags().StringVar(&taskModuleID, "module", "", "所属模块 ID")
	taskEditCmd.Flags().StringVar(&taskStoryID, "story", "", "关联所属需求 ID")
	taskEditCmd.Flags().StringVar(&taskKeywords, "keywords", "", "任务关键词")
	taskEditCmd.Flags().StringVar(&taskMailto, "mailto", "", "抄送列表")
	taskEditCmd.Flags().StringVar(&taskDeadline, "deadline", "", "截止日期")
	taskEditCmd.Flags().StringVar(&taskEstStarted, "est-started", "", "预计开始日期")
	taskEditCmd.Flags().StringVar(&taskStatus, "status", "", "任务状态")
	taskEditCmd.Flags().StringVar(&taskComment, "comment", "", "修改备注说明")

	taskStartCmd.Flags().StringVar(&taskID, "id", "", "要开始的任务 ID (必填)")
	taskStartCmd.Flags().StringVar(&taskRealStarted, "real-started", "", "实际开始时间 (格式: YYYY-MM-DD HH:MM:SS)")
	taskStartCmd.Flags().StringVar(&taskLeft, "left", "", "预计剩余工时")
	taskStartCmd.Flags().StringVar(&taskConsumed, "consumed", "", "最初消耗工时")
	taskStartCmd.Flags().StringVar(&taskComment, "comment", "", "开始任务备注")

	taskPauseCmd.Flags().StringVar(&taskID, "id", "", "要暂停的任务 ID (必填)")
	taskPauseCmd.Flags().StringVar(&taskComment, "comment", "", "暂停备注说明")

	taskRestartCmd.Flags().StringVar(&taskID, "id", "", "要重启的任务 ID (必填)")
	taskRestartCmd.Flags().StringVar(&taskLeft, "left", "", "预计剩余工时")
	taskRestartCmd.Flags().StringVar(&taskComment, "comment", "", "重启备注说明")

	taskFinishParamsCmd.Flags().StringVar(&taskID, "id", "", "要完成的任务 ID (必填)")

	taskFinishCmd.Flags().StringVar(&taskID, "id", "", "要完成的任务 ID (必填)")
	taskFinishCmd.Flags().StringVar(&taskReal, "real", "1.0", "实际消耗工时小时数 (例如: 4.0)")
	taskFinishCmd.Flags().StringVar(&taskComment, "comment", "", "完成备注说明")
	taskFinishCmd.Flags().StringVar(&taskFinishedDate, "finished-date", "", "实际完成日期与时间 (格式: YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS)")

	taskCloseCmd.Flags().StringVar(&taskID, "id", "", "要关闭的任务 ID (必填)")
	taskCloseCmd.Flags().StringVar(&taskComment, "comment", "", "关闭备注说明")

	taskCancelCmd.Flags().StringVar(&taskID, "id", "", "要取消的任务 ID (必填)")
	taskCancelCmd.Flags().StringVar(&taskComment, "comment", "", "取消备注说明")

	taskActivateCmd.Flags().StringVar(&taskID, "id", "", "要激活的任务 ID (必填)")
	taskActivateCmd.Flags().StringVar(&taskLeft, "left", "", "剩余工时")
	taskActivateCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号")
	taskActivateCmd.Flags().StringVar(&taskComment, "comment", "", "激活备注说明")

	taskAssignCmd.Flags().StringVar(&taskID, "id", "", "要指派的任务 ID (必填)")
	taskAssignCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号 (必填)")
	taskAssignCmd.Flags().StringVar(&taskLeft, "left", "", "剩余工时")
	taskAssignCmd.Flags().StringVar(&taskConsumed, "consumed", "", "已消耗工时")
	taskAssignCmd.Flags().StringVar(&taskComment, "comment", "", "指派备注说明")

	taskDeleteCmd.Flags().StringVar(&taskID, "id", "", "要删除的任务 ID (必填)")
	taskDeleteCmd.Flags().StringVar(&taskProjectID, "project", "0", "所属项目 ID 或 执行/迭代 ID")

	taskRestoreCmd.Flags().StringVar(&taskID, "id", "", "要恢复的任务 ID (必填)")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskParamsCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskEditCmd)
	taskCmd.AddCommand(taskStartCmd)
	taskCmd.AddCommand(taskPauseCmd)
	taskCmd.AddCommand(taskRestartCmd)
	taskCmd.AddCommand(taskFinishParamsCmd)
	taskCmd.AddCommand(taskFinishCmd)
	taskCmd.AddCommand(taskCloseCmd)
	taskCmd.AddCommand(taskCancelCmd)
	taskCmd.AddCommand(taskActivateCmd)
	taskCmd.AddCommand(taskAssignCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskRestoreCmd)
}
