package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var myCmd = &cobra.Command{
	Use:   "my",
	Short: "个人工作台与待办看板（我的任务、Bug、待办、需求、项目、动态）",
	Long:  "快速查看当前登录用户的专属工作项，包括指派给我的任务、Bug、待办事项、参与的项目以及活动流，支持多状态类型过滤。",
}

var (
	myTaskType       string
	myTaskOrderBy    string
	myBugType        string
	myBugOrderBy     string
	myStoryType      string
	myStoryOrderBy   string
	myTodoType       string
	myTodoStatus     string
	myProjectType    string
	myProjectOrderBy string
	myDynamicType    string
)

var myTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "查看我的任务列表（支持指派给我、我创建、已完成、未完成等筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myTaskType != "" {
			params.Set("type", myTaskType)
		}
		if myTaskOrderBy != "" {
			params.Set("orderBy", myTaskOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyTasks(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var myBugCmd = &cobra.Command{
	Use:   "bug",
	Short: "查看我的缺陷 (Bug) 列表（支持指派给我、我创建、我解决等筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myBugType != "" {
			params.Set("type", myBugType)
		}
		if myBugOrderBy != "" {
			params.Set("orderBy", myBugOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyBugs(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var myTodoCmd = &cobra.Command{
	Use:   "todo",
	Short: "查看我的日程待办列表（支持按周期与状态筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myTodoType != "" {
			params.Set("type", myTodoType)
		}
		if myTodoStatus != "" {
			params.Set("status", myTodoStatus)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyTodos(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var myStoryCmd = &cobra.Command{
	Use:   "story",
	Short: "查看我的需求 / 故事列表（支持指派给我、我创建、我评审等筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myStoryType != "" {
			params.Set("type", myStoryType)
		}
		if myStoryOrderBy != "" {
			params.Set("orderBy", myStoryOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyStories(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var myProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "查看我参与的项目列表（支持进行中、未开始、已挂起、已关闭等筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myProjectType != "" {
			params.Set("status", myProjectType)
		}
		if myProjectOrderBy != "" {
			params.Set("orderBy", myProjectOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyProjects(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var myDynamicCmd = &cobra.Command{
	Use:   "dynamic",
	Short: "查看我的最新动态 / 活动流",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if myDynamicType != "" {
			params.Set("type", myDynamicType)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.MyDynamics(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	myTaskCmd.Flags().StringVar(&myTaskType, "type", "assignedTo", "任务筛选类型: assignedTo (指派给我), openedBy (我创建的), finishedBy (我完成的), closedBy (我关闭的), canceledBy (我取消的), assignedBy (我指派的), undone (未完成), done (已完成)")
	myTaskCmd.Flags().StringVar(&myTaskOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, pri_asc, deadline_asc, status_asc)")
	addPaginationFlags(myTaskCmd)

	myBugCmd.Flags().StringVar(&myBugType, "type", "assignedTo", "Bug 筛选类型: assignedTo (指派给我), openedBy (我创建的), resolvedBy (我解决的), closedBy (我关闭的), assignedBy (我指派的)")
	myBugCmd.Flags().StringVar(&myBugOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, pri_asc, severity_desc, openedDate_desc)")
	addPaginationFlags(myBugCmd)

	myTodoCmd.Flags().StringVar(&myTodoType, "type", "all", "待办周期范围: all (全部), today (今天), thisWeek (本周), lastWeek (上周), thisMonth (本月), lastMonth (上月), before (逾期未完), future (将来待定)")
	myTodoCmd.Flags().StringVar(&myTodoStatus, "status", "all", "待办状态过滤: all (全部), wait (未开始), doing (进行中), done (已完成), closed (已关闭)")
	addPaginationFlags(myTodoCmd)

	myStoryCmd.Flags().StringVar(&myStoryType, "type", "assignedTo", "需求筛选类型: assignedTo (指派给我), openedBy (我创建的), reviewedBy (由我评审), closedBy (我关闭的), assignedBy (我指派的)")
	myStoryCmd.Flags().StringVar(&myStoryOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, pri_asc, stage_asc)")
	addPaginationFlags(myStoryCmd)

	myProjectCmd.Flags().StringVar(&myProjectType, "status", "all", "项目状态过滤: all (全部), doing (进行中), wait (未开始), suspended (已挂起), closed (已关闭), undone (未完成)")
	myProjectCmd.Flags().StringVar(&myProjectOrderBy, "order-by", "order_desc", "排序字段 (例如: order_desc, id_desc, begin_desc, end_desc)")
	addPaginationFlags(myProjectCmd)

	myDynamicCmd.Flags().StringVar(&myDynamicType, "type", "today", "动态周期范围: today (今天), yesterday (昨天), thisWeek (本周), lastWeek (上周), thisMonth (本月), all (全部)")
	addPaginationFlags(myDynamicCmd)

	myCmd.AddCommand(myTaskCmd)
	myCmd.AddCommand(myBugCmd)
	myCmd.AddCommand(myTodoCmd)
	myCmd.AddCommand(myStoryCmd)
	myCmd.AddCommand(myProjectCmd)
	myCmd.AddCommand(myDynamicCmd)
}
