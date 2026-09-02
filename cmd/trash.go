package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var trashCmd = &cobra.Command{
	Use:   "trash",
	Short: "管理禅道系统回收站（查看已删除对象、恢复、隐藏）",
	Long:  "支持查询系统回收站中各模块（task, bug, story, project, product, user, todo, doc, etc.）被删除的对象列表，并支持通过 action ID 进行恢复或隐藏。",
}

var (
	trashType     string
	trashOrderBy  string
	trashActionID string
)

var trashListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看回收站中的对象列表（支持按对象类型与排序筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if trashType != "" {
			params.Set("type", trashType)
		}
		if trashOrderBy != "" {
			params.Set("orderBy", trashOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.TrashList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var trashRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站恢复指定的删除记录（通过 action ID）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if trashActionID == "" {
			return fmt.Errorf("--action-id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TrashRestore(ctx, trashActionID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var trashHideOneCmd = &cobra.Command{
	Use:   "hide-one",
	Short: "在回收站中隐藏指定的删除记录（不再显示在回收站列表中）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if trashActionID == "" {
			return fmt.Errorf("--action-id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TrashHideOne(ctx, trashActionID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var trashHideAllCmd = &cobra.Command{
	Use:   "hide-all",
	Short: "清空/隐藏回收站中的全部删除记录",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.TrashHideAll(ctx)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	trashListCmd.Flags().StringVar(&trashType, "type", "all", "对象类型过滤: all, task, bug, story, project, execution, product, user, todo, doc 等")
	trashListCmd.Flags().StringVar(&trashOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, id_asc, date_desc, date_asc)")
	addPaginationFlags(trashListCmd)

	trashRestoreCmd.Flags().StringVar(&trashActionID, "action-id", "", "要恢复的回收站动作 ID (必填)")
	trashHideOneCmd.Flags().StringVar(&trashActionID, "action-id", "", "要隐藏的回收站动作 ID (必填)")

	trashCmd.AddCommand(trashListCmd)
	trashCmd.AddCommand(trashRestoreCmd)
	trashCmd.AddCommand(trashHideOneCmd)
	trashCmd.AddCommand(trashHideAllCmd)

	RootCmd.AddCommand(trashCmd)
}
