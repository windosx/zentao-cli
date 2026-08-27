package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "管理禅道项目与执行",
	Long:  "查询禅道项目/迭代列表或创建新项目/执行，支持按状态（进行中、未完成、未开始、已挂起、已关闭、全部）、所属产品、项目集等筛选。",
}

var (
	projectStatus    string
	projectOrderBy   string
	projectProductID string
	projectProgramID string
	projectName      string
	projectCode      string
	projectBegin     string
	projectEnd       string
	projectDays      string
	projectTeam      string
	projectType      string
	projectPM        string
	projectPO        string
	projectQD        string
	projectRD        string
	projectACL       string
	projectDesc      string
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询项目列表（支持按状态、产品、项目集筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectStatus != "" {
			params.Set("browseType", projectStatus)
			params.Set("status", projectStatus)
		}
		if projectOrderBy != "" {
			params.Set("orderBy", projectOrderBy)
		}
		if projectProductID != "" {
			params.Set("productID", projectProductID)
		}
		if projectProgramID != "" {
			params.Set("programID", projectProgramID)
		}

		data, err := client.ProjectList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取创建项目所需的元数据字典（项目集、可用产品列表等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProjectCreateParams(ctx, projectProgramID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectAddCmd = &cobra.Command{
	Use:   "add",
	Short: "创建新项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectName == "" {
			return fmt.Errorf("--name 是必填参数")
		}
		if projectCode == "" {
			return fmt.Errorf("--code 是必填参数")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {projectName},
			"code": {projectCode},
		}
		if projectBegin != "" {
			params.Set("begin", projectBegin)
		}
		if projectEnd != "" {
			params.Set("end", projectEnd)
		}
		if projectDays != "" {
			params.Set("days", projectDays)
		}
		if projectTeam != "" {
			params.Set("team", projectTeam)
		}
		if projectType != "" {
			params.Set("type", projectType)
		}
		if projectPM != "" {
			params.Set("PM", projectPM)
		}
		if projectPO != "" {
			params.Set("PO", projectPO)
		}
		if projectQD != "" {
			params.Set("QD", projectQD)
		}
		if projectRD != "" {
			params.Set("RD", projectRD)
		}
		if projectACL != "" {
			params.Set("acl", projectACL)
		}
		if projectDesc != "" {
			params.Set("desc", projectDesc)
		}
		if projectProductID != "" {
			params.Set("products[0]", projectProductID)
		}

		data, err := client.ProjectAdd(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	projectListCmd.Flags().StringVar(&projectStatus, "status", "doing", "项目状态过滤: doing (进行中), undone (未完成/全部进行中的状态), all (全部项目), wait (未开始), suspended (已挂起), closed (已关闭)")
	projectListCmd.Flags().StringVar(&projectProductID, "product", "", "关联所属产品 ID")
	projectListCmd.Flags().StringVar(&projectProgramID, "program", "0", "按项目集 ID 过滤 (0 为全部)")
	projectListCmd.Flags().StringVar(&projectOrderBy, "order-by", "order_desc", "排序字段 (例如: order_desc, order_asc, id_desc, id_asc, begin_desc, end_desc)")

	projectParamsCmd.Flags().StringVar(&projectProgramID, "program", "0", "按项目集 ID 过滤 (0 为全部)")

	projectAddCmd.Flags().StringVar(&projectName, "name", "", "项目名称 (必填)")
	projectAddCmd.Flags().StringVar(&projectCode, "code", "", "项目代号 / 英文缩写 (必填)")
	projectAddCmd.Flags().StringVar(&projectBegin, "begin", "", "计划开始日期 (格式: YYYY-MM-DD)")
	projectAddCmd.Flags().StringVar(&projectEnd, "end", "", "计划结束日期 (格式: YYYY-MM-DD)")
	projectAddCmd.Flags().StringVar(&projectDays, "days", "", "可用工日天数")
	projectAddCmd.Flags().StringVar(&projectTeam, "team", "", "团队名称")
	projectAddCmd.Flags().StringVar(&projectType, "type", "sprint", "项目模型类型: sprint (敏捷迭代), waterfall (瀑布开发)")
	projectAddCmd.Flags().StringVar(&projectPM, "pm", "", "项目经理 (PM) 账号")
	projectAddCmd.Flags().StringVar(&projectPO, "po", "", "产品负责人 (PO) 账号")
	projectAddCmd.Flags().StringVar(&projectQD, "qd", "", "测试负责人 (QD) 账号")
	projectAddCmd.Flags().StringVar(&projectRD, "rd", "", "开发负责人 (RD) 账号")
	projectAddCmd.Flags().StringVar(&projectACL, "acl", "open", "访问权限控制: open (公开), custom (自定义), private (私有)")
	projectAddCmd.Flags().StringVar(&projectDesc, "desc", "", "项目详细描述说明")
	projectAddCmd.Flags().StringVar(&projectProductID, "product", "", "关联所属产品 ID")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectParamsCmd)
	projectCmd.AddCommand(projectAddCmd)
}
