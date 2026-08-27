package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var deptCmd = &cobra.Command{
	Use:   "dept",
	Short: "管理禅道部门",
	Long:  "查询公司部门层级树、在指定父部门下创建子部门、修改部门或删除部门。",
}

var (
	deptID       string
	deptParentID string
	deptName     string
	deptManager  string
	deptNames    []string
)

var deptListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询部门层级结构列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if deptParentID != "" {
			params.Set("deptID", deptParentID)
		}

		data, err := client.DeptList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var deptCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "在父部门下添加子部门",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(deptNames) == 0 {
			return fmt.Errorf("--name 是必填参数（可多次指定添加多个部门）")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		parent := deptParentID
		if parent == "" {
			parent = "0"
		}

		params := zentao.Params{}
		params.Set("parentDeptID", parent)
		for i, name := range deptNames {
			params.Set(fmt.Sprintf("depts[%d]", i), name)
		}

		data, err := client.DeptAdd(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var deptEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定部门信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if deptID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if deptName != "" {
			params.Set("name", deptName)
		}
		if deptParentID != "" {
			params.Set("parent", deptParentID)
		}
		if deptManager != "" {
			params.Set("manager", deptManager)
		}

		data, err := client.DeptEdit(ctx, deptID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var deptDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定部门",
	RunE: func(cmd *cobra.Command, args []string) error {
		if deptID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.DeptDelete(ctx, deptID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	deptListCmd.Flags().StringVar(&deptParentID, "parent", "", "父部门 ID (默认: 顶级/根部门)")

	deptCreateCmd.Flags().StringVar(&deptParentID, "parent", "0", "父部门 ID")
	deptCreateCmd.Flags().StringSliceVar(&deptNames, "name", nil, "要添加的子部门名称列表 (可多次指定)")

	deptEditCmd.Flags().StringVar(&deptID, "id", "", "要修改的部门 ID (必填)")
	deptEditCmd.Flags().StringVar(&deptName, "name", "", "部门名称")
	deptEditCmd.Flags().StringVar(&deptParentID, "parent", "", "上级父部门 ID")
	deptEditCmd.Flags().StringVar(&deptManager, "manager", "", "部门负责人账号")

	deptDeleteCmd.Flags().StringVar(&deptID, "id", "", "要删除的部门 ID (必填)")

	deptCmd.AddCommand(deptListCmd)
	deptCmd.AddCommand(deptCreateCmd)
	deptCmd.AddCommand(deptEditCmd)
	deptCmd.AddCommand(deptDeleteCmd)
}
