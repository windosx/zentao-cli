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
	Long:  "查询公司部门层级树或在指定父部门下创建子部门。",
}

var (
	deptParentID string
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

var deptAddCmd = &cobra.Command{
	Use:   "add",
	Short: "在父部门下添加子部门",
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

func init() {
	deptListCmd.Flags().StringVar(&deptParentID, "parent", "", "父部门 ID (默认: 顶级/根部门)")
	deptAddCmd.Flags().StringVar(&deptParentID, "parent", "0", "父部门 ID")
	deptAddCmd.Flags().StringSliceVar(&deptNames, "name", nil, "要添加的子部门名称列表 (可多次指定)")

	deptCmd.AddCommand(deptListCmd)
	deptCmd.AddCommand(deptAddCmd)
}
