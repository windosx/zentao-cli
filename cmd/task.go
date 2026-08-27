package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "管理禅道任务",
	Long:  "查询项目/执行下的任务列表、获取创建任务元数据、创建新任务或标记完成任务，支持状态、类型、优先级、指派人多维度过滤。",
}

var (
	taskProjectID    string
	taskID           string
	taskStatus       string
	taskOrderBy      string
	taskName         string
	taskType         string
	taskPri          string
	taskEstimate     string
	taskAssignedTo   string
	taskDesc         string
	taskModuleID     string
	taskStoryID      string
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

		ctx := context.Background()
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

		data, err := client.TaskList(ctx, params)
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

		ctx := context.Background()
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

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"project": {taskProjectID},
			"name":    {taskName},
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

		data, err := client.TaskCreate(ctx, params)
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

		ctx := context.Background()
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
	Short: "完成并关闭任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := context.Background()
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

var taskDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := context.Background()
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

func init() {
	taskListCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 ID 或 执行/迭代 ID (必填)")
	taskListCmd.Flags().StringVar(&taskStatus, "status", "all", "任务状态过滤: all (全部), undone/unclosed (未完成), wait (未开始), doing (进行中), done (已完成), pause (已暂停), cancel (已取消), closed (已关闭), needconfirm (需求变动待确认)")
	taskListCmd.Flags().StringVar(&taskOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, id_asc, pri_asc, deadline_asc, estimate_desc, consumed_desc)")

	taskParamsCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 / 执行 ID (必填)")

	taskCreateCmd.Flags().StringVar(&taskProjectID, "project", "", "所属项目 / 执行 ID (必填)")
	taskCreateCmd.Flags().StringVar(&taskName, "name", "", "任务名称 (必填)")
	taskCreateCmd.Flags().StringVar(&taskType, "type", "devel", "任务类型 (design: 设计, devel: 开发, test: 测试, study: 研究, discuss: 讨论, ui: 界面, affair: 事务, misc: 其他)")
	taskCreateCmd.Flags().StringVar(&taskPri, "pri", "3", "任务优先级 (1=最高, 2=高, 3=中, 4=低)")
	taskCreateCmd.Flags().StringVar(&taskEstimate, "estimate", "", "预计工时 (例如: 4.5)")
	taskCreateCmd.Flags().StringVar(&taskAssignedTo, "assigned-to", "", "指派给的用户账号")
	taskCreateCmd.Flags().StringVar(&taskDesc, "desc", "", "任务详细描述说明")
	taskCreateCmd.Flags().StringVar(&taskModuleID, "module", "0", "所属模块 ID")
	taskCreateCmd.Flags().StringVar(&taskStoryID, "story", "0", "关联所属需求/故事 ID")

	taskFinishCmd.Flags().StringVar(&taskID, "id", "", "要完成的任务 ID (必填)")
	taskFinishCmd.Flags().StringVar(&taskReal, "real", "1.0", "实际消耗工时小时数 (例如: 4.0)")
	taskFinishCmd.Flags().StringVar(&taskComment, "comment", "", "完成备注说明")
	taskFinishCmd.Flags().StringVar(&taskFinishedDate, "finished-date", "", "实际完成日期与时间 (格式: YYYY-MM-DD 或 YYYY-MM-DD HH:MM:SS)")

	taskFinishParamsCmd.Flags().StringVar(&taskID, "id", "", "要完成的任务 ID (必填)")

	taskDeleteCmd.Flags().StringVar(&taskID, "id", "", "要删除的任务 ID (必填)")
	taskDeleteCmd.Flags().StringVar(&taskProjectID, "project", "0", "所属项目 ID 或 执行/迭代 ID")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskParamsCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskFinishParamsCmd)
	taskCmd.AddCommand(taskFinishCmd)
	taskCmd.AddCommand(taskDeleteCmd)
}
