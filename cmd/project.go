package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "管理禅道项目与执行",
	Long:  "查询禅道项目/迭代列表、查看项目详情、创建新项目/执行、编辑项目信息或执行生命周期操作（开始、挂起、激活、关闭、删除）。",
}

var (
	projectID        string
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
	projectPri       string
	projectACL       string
	projectWhitelist string
	projectDesc      string
	projectComment   string
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询项目列表（支持按状态、产品、项目集筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
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
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ProjectList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定项目的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProjectView(ctx, projectID)
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
		ctx := cmd.Context()
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

var projectCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "创建新项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectName == "" {
			return fmt.Errorf("--name 是必填参数")
		}
		if projectCode == "" {
			return fmt.Errorf("--code 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {projectName},
			"code": {projectCode},
		}
		if projectProgramID != "" {
			params.Set("program", projectProgramID)
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
		if projectPri != "" {
			params.Set("pri", projectPri)
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
		if projectWhitelist != "" {
			params.Set("whitelist", projectWhitelist)
		}
		if projectStatus != "" {
			params.Set("status", projectStatus)
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

var projectEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定项目的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectName != "" {
			params.Set("name", projectName)
		}
		if projectCode != "" {
			params.Set("code", projectCode)
		}
		if projectProgramID != "" {
			params.Set("program", projectProgramID)
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
		if projectPri != "" {
			params.Set("pri", projectPri)
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
		if projectWhitelist != "" {
			params.Set("whitelist", projectWhitelist)
		}
		if projectStatus != "" {
			params.Set("status", projectStatus)
		}
		if projectDesc != "" {
			params.Set("desc", projectDesc)
		}
		if projectProductID != "" {
			params.Set("products[0]", projectProductID)
		}
		if projectComment != "" {
			params.Set("comment", projectComment)
		}

		data, err := client.ProjectEdit(ctx, projectID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectStartCmd = &cobra.Command{
	Use:   "start",
	Short: "开始进行项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectComment != "" {
			params.Set("comment", projectComment)
		}

		data, err := client.ProjectStart(ctx, projectID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectSuspendCmd = &cobra.Command{
	Use:   "suspend",
	Short: "挂起项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectComment != "" {
			params.Set("comment", projectComment)
		}

		data, err := client.ProjectSuspend(ctx, projectID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已挂起或已关闭的项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectComment != "" {
			params.Set("comment", projectComment)
		}

		data, err := client.ProjectActivate(ctx, projectID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if projectComment != "" {
			params.Set("comment", projectComment)
		}

		data, err := client.ProjectClose(ctx, projectID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProjectDelete(ctx, projectID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var projectRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的项目",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "project", projectID)
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
	addPaginationFlags(projectListCmd)

	projectViewCmd.Flags().StringVar(&projectID, "id", "", "要查看的项目 ID (必填)")

	projectParamsCmd.Flags().StringVar(&projectProgramID, "program", "0", "按项目集 ID 过滤 (0 为全部)")

	projectCreateCmd.Flags().StringVar(&projectName, "name", "", "项目名称 (必填)")
	projectCreateCmd.Flags().StringVar(&projectCode, "code", "", "项目代号 / 英文缩写 (必填)")
	projectCreateCmd.Flags().StringVar(&projectProgramID, "program", "0", "所属项目集 ID")
	projectCreateCmd.Flags().StringVar(&projectBegin, "begin", "", "计划开始日期 (格式: YYYY-MM-DD)")
	projectCreateCmd.Flags().StringVar(&projectEnd, "end", "", "计划结束日期 (格式: YYYY-MM-DD)")
	projectCreateCmd.Flags().StringVar(&projectDays, "days", "", "可用工日天数")
	projectCreateCmd.Flags().StringVar(&projectTeam, "team", "", "团队名称")
	projectCreateCmd.Flags().StringVar(&projectType, "type", "sprint", "项目模型类型: sprint (敏捷迭代), waterfall (瀑布开发)")
	projectCreateCmd.Flags().StringVar(&projectPri, "pri", "3", "项目优先级")
	projectCreateCmd.Flags().StringVar(&projectPM, "pm", "", "项目经理 (PM) 账号")
	projectCreateCmd.Flags().StringVar(&projectPO, "po", "", "产品负责人 (PO) 账号")
	projectCreateCmd.Flags().StringVar(&projectQD, "qd", "", "测试负责人 (QD) 账号")
	projectCreateCmd.Flags().StringVar(&projectRD, "rd", "", "开发负责人 (RD) 账号")
	projectCreateCmd.Flags().StringVar(&projectACL, "acl", "open", "访问权限控制: open (公开), custom (自定义), private (私有)")
	projectCreateCmd.Flags().StringVar(&projectWhitelist, "whitelist", "", "分组白名单列表 (逗号分隔)")
	projectCreateCmd.Flags().StringVar(&projectStatus, "status", "wait", "初始状态 (wait, doing)")
	projectCreateCmd.Flags().StringVar(&projectDesc, "desc", "", "项目详细描述说明")
	projectCreateCmd.Flags().StringVar(&projectProductID, "product", "", "关联所属产品 ID")

	projectEditCmd.Flags().StringVar(&projectID, "id", "", "要修改的项目 ID (必填)")
	projectEditCmd.Flags().StringVar(&projectName, "name", "", "项目名称")
	projectEditCmd.Flags().StringVar(&projectCode, "code", "", "项目代号")
	projectEditCmd.Flags().StringVar(&projectProgramID, "program", "", "所属项目集 ID")
	projectEditCmd.Flags().StringVar(&projectBegin, "begin", "", "计划开始日期")
	projectEditCmd.Flags().StringVar(&projectEnd, "end", "", "计划结束日期")
	projectEditCmd.Flags().StringVar(&projectDays, "days", "", "可用工日天数")
	projectEditCmd.Flags().StringVar(&projectTeam, "team", "", "团队名称")
	projectEditCmd.Flags().StringVar(&projectType, "type", "", "项目类型")
	projectEditCmd.Flags().StringVar(&projectPri, "pri", "", "项目优先级")
	projectEditCmd.Flags().StringVar(&projectPM, "pm", "", "项目经理 (PM) 账号")
	projectEditCmd.Flags().StringVar(&projectPO, "po", "", "产品负责人 (PO) 账号")
	projectEditCmd.Flags().StringVar(&projectQD, "qd", "", "测试负责人 (QD) 账号")
	projectEditCmd.Flags().StringVar(&projectRD, "rd", "", "开发负责人 (RD) 账号")
	projectEditCmd.Flags().StringVar(&projectACL, "acl", "", "访问权限控制")
	projectEditCmd.Flags().StringVar(&projectWhitelist, "whitelist", "", "白名单列表")
	projectEditCmd.Flags().StringVar(&projectStatus, "status", "", "项目状态")
	projectEditCmd.Flags().StringVar(&projectDesc, "desc", "", "项目描述说明")
	projectEditCmd.Flags().StringVar(&projectProductID, "product", "", "关联产品 ID")
	projectEditCmd.Flags().StringVar(&projectComment, "comment", "", "修改备注说明")

	projectStartCmd.Flags().StringVar(&projectID, "id", "", "要开始的项目 ID (必填)")
	projectStartCmd.Flags().StringVar(&projectComment, "comment", "", "开始备注说明")

	projectSuspendCmd.Flags().StringVar(&projectID, "id", "", "要挂起的项目 ID (必填)")
	projectSuspendCmd.Flags().StringVar(&projectComment, "comment", "", "挂起备注说明")

	projectActivateCmd.Flags().StringVar(&projectID, "id", "", "要激活的项目 ID (必填)")
	projectActivateCmd.Flags().StringVar(&projectComment, "comment", "", "激活备注说明")

	projectCloseCmd.Flags().StringVar(&projectID, "id", "", "要关闭的项目 ID (必填)")
	projectCloseCmd.Flags().StringVar(&projectComment, "comment", "", "关闭备注说明")

	projectDeleteCmd.Flags().StringVar(&projectID, "id", "", "要删除的项目 ID (必填)")
	projectRestoreCmd.Flags().StringVar(&projectID, "id", "", "要恢复的项目 ID (必填)")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectViewCmd)
	projectCmd.AddCommand(projectParamsCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectEditCmd)
	projectCmd.AddCommand(projectStartCmd)
	projectCmd.AddCommand(projectSuspendCmd)
	projectCmd.AddCommand(projectActivateCmd)
	projectCmd.AddCommand(projectCloseCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectRestoreCmd)
}
