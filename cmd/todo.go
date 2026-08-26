package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "管理个人待办事项 (Todo)",
	Long:  "查看待办、创建自定义待办、开始待办、完成待办或关闭待办。",
}

var (
	todoType   string
	todoStatus string
	todoID     string
	todoName   string
	todoDate   string
	todoBegin  string
	todoEnd    string
	todoPri    string
	todoDesc   string
)

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询个人待办列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
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

		data, err := client.TodoList(ctx, params)
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

		ctx := context.Background()
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
		params.Set("type", "custom")

		data, err := client.TodoCreate(ctx, params)
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

		ctx := context.Background()
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

		ctx := context.Background()
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

		ctx := context.Background()
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

func init() {
	todoListCmd.Flags().StringVar(&todoType, "type", "all", "待办周期: today (今天), thisWeek (本周), lastWeek (上周), thisMonth (本月), before (逾期待办), future (待定), all (全部)")
	todoListCmd.Flags().StringVar(&todoStatus, "status", "all", "待办状态: all (全部), wait (未开始), doing (进行中), done (已完成), closed (已关闭)")

	todoCreateCmd.Flags().StringVar(&todoName, "name", "", "待办名称 / 描述 (必填)")
	todoCreateCmd.Flags().StringVar(&todoDate, "date", "", "待办日期 (格式: YYYY-MM-DD，默认今天)")
	todoCreateCmd.Flags().StringVar(&todoBegin, "begin", "", "开始时间 (例如: 0900)")
	todoCreateCmd.Flags().StringVar(&todoEnd, "end", "", "结束时间 (例如: 1800)")
	todoCreateCmd.Flags().StringVar(&todoPri, "pri", "3", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	todoCreateCmd.Flags().StringVar(&todoDesc, "desc", "", "待办详细说明")

	todoStartCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoFinishCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")
	todoCloseCmd.Flags().StringVar(&todoID, "id", "", "待办 ID (必填)")

	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoCreateCmd)
	todoCmd.AddCommand(todoStartCmd)
	todoCmd.AddCommand(todoFinishCmd)
	todoCmd.AddCommand(todoCloseCmd)
}
