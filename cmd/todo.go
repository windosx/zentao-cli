package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "管理个人待办事项 (Todo)",
	Long:  "查看待办列表、查看待办详情、创建待办（支持自定义/Bug/任务/需求关联）、编辑待办或执行生命周期操作（开始、完成、关闭、激活、指派、删除）。",
}

var (
	todoType       string
	todoStatus     string
	todoOrderBy    string
	todoID         string
	todoName       string
	todoDate       string
	todoBegin      string
	todoEnd        string
	todoPri        string
	todoDesc       string
	todoIDValue    string
	todoPrivate    string
	todoAssignedTo string
)

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询个人待办列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if todoType != "" {
			params.Set("type", todoType)
		}
		if todoStatus != "" {
			params.Set("status", todoStatus)
		}
		if todoOrderBy != "" {
			params.Set("orderBy", todoOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.TodoList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定待办事项的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoView(ctx, todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新的待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoName == "" {
			return fmt.Errorf("--name 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {todoName},
		}
		if todoDate != "" {
			params.Set("date", todoDate)
		}
		if todoBegin != "" {
			params.Set("begin", todoBegin)
		}
		if todoEnd != "" {
			params.Set("end", todoEnd)
		}
		if todoPri != "" {
			params.Set("pri", todoPri)
		}
		if todoDesc != "" {
			params.Set("desc", todoDesc)
		}
		if todoType != "" {
			params.Set("type", todoType)
		}
		if todoIDValue != "" {
			params.Set("idvalue", todoIDValue)
		}
		if todoPrivate != "" {
			params.Set("private", todoPrivate)
		}

		data, err := client.TodoCreate(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定待办事项的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if todoName != "" {
			params.Set("name", todoName)
		}
		if todoDate != "" {
			params.Set("date", todoDate)
		}
		if todoBegin != "" {
			params.Set("begin", todoBegin)
		}
		if todoEnd != "" {
			params.Set("end", todoEnd)
		}
		if todoPri != "" {
			params.Set("pri", todoPri)
		}
		if todoDesc != "" {
			params.Set("desc", todoDesc)
		}
		if todoType != "" {
			params.Set("type", todoType)
		}
		if todoIDValue != "" {
			params.Set("idvalue", todoIDValue)
		}
		if todoPrivate != "" {
			params.Set("private", todoPrivate)
		}
		if todoStatus != "" {
			params.Set("status", todoStatus)
		}

		data, err := client.TodoEdit(ctx, todoID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoStartCmd = &cobra.Command{
	Use:   "start",
	Short: "开始进行中的待办",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoStart(ctx, todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "标记完成待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoFinish(ctx, todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoClose(ctx, todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已完成或已关闭的待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoActivate(ctx, todoID, zentao.Params{})
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "指派待办事项给其他用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if todoAssignedTo == "" {
			return fmt.Errorf("--assigned-to 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{"assignedTo": {todoAssignedTo}}
		data, err := client.TodoAssign(ctx, todoID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TodoDelete(ctx, todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var todoRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的待办事项",
	RunE: func(cmd *cobra.Command, args []string) error {
		if todoID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "todo", todoID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	todoListCmd.Flags().StringVar(&todoType, "type", "all", "待办周期: today (今天), thisWeek (本周), lastWeek (上周), thisMonth (本月), before (逾期待办), future (待定), all (全部)")
	todoListCmd.Flags().StringVar(&todoStatus, "status", "all", "待办状态: all (全部), wait (未开始), doing (进行中), done (已完成), closed (已关闭)")
	todoListCmd.Flags().StringVar(&todoOrderBy, "order-by", "date_desc,status,id_desc", "排序字段")
	addPaginationFlags(todoListCmd)

	todoViewCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")

	todoCreateCmd.Flags().StringVar(&todoName, "name", "", "待办名称 / 描述 (必填)")
	todoCreateCmd.Flags().StringVar(&todoDate, "date", "", "待办日期 (格式: YYYY-MM-DD，默认今天)")
	todoCreateCmd.Flags().StringVar(&todoBegin, "begin", "", "开始时间 (例如: 0900)")
	todoCreateCmd.Flags().StringVar(&todoEnd, "end", "", "结束时间 (例如: 1800)")
	todoCreateCmd.Flags().StringVar(&todoType, "type", "custom", "待办类型 (custom: 自定义, task: 任务, bug: Bug, story: 需求)")
	todoCreateCmd.Flags().StringVar(&todoIDValue, "idvalue", "0", "关联的任务/Bug/需求 ID")
	todoCreateCmd.Flags().StringVar(&todoPri, "pri", "3", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	todoCreateCmd.Flags().StringVar(&todoDesc, "desc", "", "待办详细说明")
	todoCreateCmd.Flags().StringVar(&todoPrivate, "private", "0", "是否私有待办 (0=否, 1=是)")

	todoEditCmd.Flags().StringVar(&todoID, "id", "", "要修改的待办 ID (必填)")
	todoEditCmd.Flags().StringVar(&todoName, "name", "", "待办名称 / 描述")
	todoEditCmd.Flags().StringVar(&todoDate, "date", "", "待办日期")
	todoEditCmd.Flags().StringVar(&todoBegin, "begin", "", "开始时间")
	todoEditCmd.Flags().StringVar(&todoEnd, "end", "", "结束时间")
	todoEditCmd.Flags().StringVar(&todoType, "type", "", "待办类型")
	todoEditCmd.Flags().StringVar(&todoIDValue, "idvalue", "", "关联实体 ID")
	todoEditCmd.Flags().StringVar(&todoPri, "pri", "", "优先级")
	todoEditCmd.Flags().StringVar(&todoDesc, "desc", "", "待办详细说明")
	todoEditCmd.Flags().StringVar(&todoStatus, "status", "", "待办状态 (wait, doing, done, closed)")
	todoEditCmd.Flags().StringVar(&todoPrivate, "private", "", "是否私有")

	todoStartCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoFinishCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoCloseCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoActivateCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")

	todoAssignCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoAssignCmd.Flags().StringVar(&todoAssignedTo, "assigned-to", "", "指派给的用户账号 (必填)")

	todoDeleteCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoRestoreCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")

	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoViewCmd)
	todoCmd.AddCommand(todoCreateCmd)
	todoCmd.AddCommand(todoEditCmd)
	todoCmd.AddCommand(todoStartCmd)
	todoCmd.AddCommand(todoFinishCmd)
	todoCmd.AddCommand(todoCloseCmd)
	todoCmd.AddCommand(todoActivateCmd)
	todoCmd.AddCommand(todoAssignCmd)
	todoCmd.AddCommand(todoDeleteCmd)
	todoCmd.AddCommand(todoRestoreCmd)
}
