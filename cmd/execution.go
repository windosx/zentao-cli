package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var executionCmd = &cobra.Command{
	Use:     "execution",
	Aliases: []string{"exec", "iteration", "iter", "sprint"},
	Short:   "管理禅道迭代/执行/阶段",
	Long:    "查询禅道执行/迭代列表、查看执行详情、创建新迭代/执行、编辑迭代信息或执行生命周期操作（开始、挂起、激活、关闭、删除、恢复），以及查看关联的任务、需求与缺陷。",
}

var (
	executionID        string
	executionProjectID string
	executionStatus    string
	executionOrderBy   string
	executionName      string
	executionCode      string
	executionBegin     string
	executionEnd       string
	executionDays      string
	executionTeam      string
	executionType      string
	executionPri       string
	executionPM        string
	executionPO        string
	executionQD        string
	executionRD        string
	executionACL       string
	executionWhitelist string
	executionDesc      string
	executionComment   string
)

var executionListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询执行/迭代列表（支持按所属项目、状态筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionProjectID != "" {
			params.Set("project", executionProjectID)
			params.Set("projectID", executionProjectID)
		}
		if executionStatus != "" {
			params.Set("status", executionStatus)
			params.Set("browseType", executionStatus)
		}
		if executionOrderBy != "" {
			params.Set("orderBy", executionOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ExecutionList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定执行/迭代的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ExecutionView(ctx, executionID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取创建执行/迭代所需的元数据字典（所属项目、团队成员等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ExecutionCreateParams(ctx, executionProjectID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "创建新执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionName == "" {
			return fmt.Errorf("--name 是必填参数")
		}
		if executionCode == "" {
			return fmt.Errorf("--code 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {executionName},
			"code": {executionCode},
		}
		if executionProjectID != "" {
			params.Set("project", executionProjectID)
		}
		if executionBegin != "" {
			params.Set("begin", executionBegin)
		}
		if executionEnd != "" {
			params.Set("end", executionEnd)
		}
		if executionDays != "" {
			params.Set("days", executionDays)
		}
		if executionTeam != "" {
			params.Set("team", executionTeam)
		}
		if executionType != "" {
			params.Set("type", executionType)
		}
		if executionPri != "" {
			params.Set("pri", executionPri)
		}
		if executionPM != "" {
			params.Set("PM", executionPM)
		}
		if executionPO != "" {
			params.Set("PO", executionPO)
		}
		if executionQD != "" {
			params.Set("QD", executionQD)
		}
		if executionRD != "" {
			params.Set("RD", executionRD)
		}
		if executionACL != "" {
			params.Set("acl", executionACL)
		}
		if executionWhitelist != "" {
			params.Set("whitelist", executionWhitelist)
		}
		if executionDesc != "" {
			params.Set("desc", executionDesc)
		}

		data, err := client.ExecutionCreate(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "编辑执行/迭代信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionName != "" {
			params.Set("name", executionName)
		}
		if executionCode != "" {
			params.Set("code", executionCode)
		}
		if executionProjectID != "" {
			params.Set("project", executionProjectID)
		}
		if executionBegin != "" {
			params.Set("begin", executionBegin)
		}
		if executionEnd != "" {
			params.Set("end", executionEnd)
		}
		if executionDays != "" {
			params.Set("days", executionDays)
		}
		if executionTeam != "" {
			params.Set("team", executionTeam)
		}
		if executionType != "" {
			params.Set("type", executionType)
		}
		if executionPri != "" {
			params.Set("pri", executionPri)
		}
		if executionPM != "" {
			params.Set("PM", executionPM)
		}
		if executionPO != "" {
			params.Set("PO", executionPO)
		}
		if executionQD != "" {
			params.Set("QD", executionQD)
		}
		if executionRD != "" {
			params.Set("RD", executionRD)
		}
		if executionACL != "" {
			params.Set("acl", executionACL)
		}
		if executionWhitelist != "" {
			params.Set("whitelist", executionWhitelist)
		}
		if executionStatus != "" {
			params.Set("status", executionStatus)
		}
		if executionDesc != "" {
			params.Set("desc", executionDesc)
		}
		if executionComment != "" {
			params.Set("comment", executionComment)
		}

		data, err := client.ExecutionEdit(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionStartCmd = &cobra.Command{
	Use:   "start",
	Short: "开始进行执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionComment != "" {
			params.Set("comment", executionComment)
		}

		data, err := client.ExecutionStart(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionSuspendCmd = &cobra.Command{
	Use:     "suspend",
	Aliases: []string{"putoff"},
	Short:   "挂起执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionComment != "" {
			params.Set("comment", executionComment)
		}

		data, err := client.ExecutionSuspend(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活挂起或已关闭的执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionComment != "" {
			params.Set("comment", executionComment)
		}

		data, err := client.ExecutionActivate(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionComment != "" {
			params.Set("comment", executionComment)
		}

		data, err := client.ExecutionClose(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ExecutionDelete(ctx, executionID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的执行/迭代",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "execution", executionID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "查看执行/迭代下的任务列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionStatus != "" {
			params.Set("status", executionStatus)
		}
		if executionOrderBy != "" {
			params.Set("orderBy", executionOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ExecutionTask(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionStoryCmd = &cobra.Command{
	Use:   "story",
	Short: "查看执行/迭代关联的需求列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionOrderBy != "" {
			params.Set("orderBy", executionOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ExecutionStory(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var executionBugCmd = &cobra.Command{
	Use:   "bug",
	Short: "查看执行/迭代关联的缺陷列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		if executionID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if executionOrderBy != "" {
			params.Set("orderBy", executionOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ExecutionBug(ctx, executionID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	executionListCmd.Flags().StringVar(&executionProjectID, "project", "", "按所属项目 ID 筛选 (0 为不限)")
	executionListCmd.Flags().StringVar(&executionStatus, "status", "undone", "状态过滤: undone (未完成), doing (进行中), all (全部), wait (未开始), suspended (已挂起), closed (已关闭)")
	executionListCmd.Flags().StringVar(&executionOrderBy, "order-by", "order_desc", "排序字段 (例如: order_desc, order_asc, id_desc, id_asc, begin_desc, end_desc)")
	addPaginationFlags(executionListCmd)

	executionViewCmd.Flags().StringVar(&executionID, "id", "", "要查看的执行/迭代 ID (必填)")

	executionParamsCmd.Flags().StringVar(&executionProjectID, "project", "0", "按所属项目 ID 过滤 (0 为全部)")

	executionCreateCmd.Flags().StringVar(&executionProjectID, "project", "0", "所属项目 ID (0 为独立执行/迭代)")
	executionCreateCmd.Flags().StringVar(&executionName, "name", "", "执行/迭代名称 (必填)")
	executionCreateCmd.Flags().StringVar(&executionCode, "code", "", "执行/迭代代号 (必填)")
	executionCreateCmd.Flags().StringVar(&executionBegin, "begin", "", "计划开始日期 (格式: YYYY-MM-DD)")
	executionCreateCmd.Flags().StringVar(&executionEnd, "end", "", "计划结束日期 (格式: YYYY-MM-DD)")
	executionCreateCmd.Flags().StringVar(&executionDays, "days", "", "可用工作日天数")
	executionCreateCmd.Flags().StringVar(&executionTeam, "team", "", "团队名称")
	executionCreateCmd.Flags().StringVar(&executionType, "type", "sprint", "执行类型: sprint (迭代), stage (阶段), kanban (看板), ops (运维)")
	executionCreateCmd.Flags().StringVar(&executionPri, "pri", "1", "优先级 (1-4)")
	executionCreateCmd.Flags().StringVar(&executionPM, "pm", "", "项目经理/负责人账号")
	executionCreateCmd.Flags().StringVar(&executionPO, "po", "", "产品负责人账号")
	executionCreateCmd.Flags().StringVar(&executionQD, "qd", "", "测试负责人账号")
	executionCreateCmd.Flags().StringVar(&executionRD, "rd", "", "发布负责人账号")
	executionCreateCmd.Flags().StringVar(&executionACL, "acl", "open", "访问控制权限: open (公开), private (私有), custom (自定义白名单)")
	executionCreateCmd.Flags().StringVar(&executionWhitelist, "whitelist", "", "白名单用户分组或账号列表 (逗号分隔)")
	executionCreateCmd.Flags().StringVar(&executionDesc, "desc", "", "执行/迭代描述")

	executionEditCmd.Flags().StringVar(&executionID, "id", "", "要编辑的执行/迭代 ID (必填)")
	executionEditCmd.Flags().StringVar(&executionProjectID, "project", "", "所属项目 ID")
	executionEditCmd.Flags().StringVar(&executionName, "name", "", "执行/迭代名称")
	executionEditCmd.Flags().StringVar(&executionCode, "code", "", "执行/迭代代号")
	executionEditCmd.Flags().StringVar(&executionBegin, "begin", "", "计划开始日期 (格式: YYYY-MM-DD)")
	executionEditCmd.Flags().StringVar(&executionEnd, "end", "", "计划结束日期 (格式: YYYY-MM-DD)")
	executionEditCmd.Flags().StringVar(&executionDays, "days", "", "可用工作日天数")
	executionEditCmd.Flags().StringVar(&executionTeam, "team", "", "团队名称")
	executionEditCmd.Flags().StringVar(&executionType, "type", "", "执行类型: sprint, stage, kanban, ops")
	executionEditCmd.Flags().StringVar(&executionPri, "pri", "", "优先级 (1-4)")
	executionEditCmd.Flags().StringVar(&executionPM, "pm", "", "项目经理/负责人账号")
	executionEditCmd.Flags().StringVar(&executionPO, "po", "", "产品负责人账号")
	executionEditCmd.Flags().StringVar(&executionQD, "qd", "", "测试负责人账号")
	executionEditCmd.Flags().StringVar(&executionRD, "rd", "", "发布负责人账号")
	executionEditCmd.Flags().StringVar(&executionACL, "acl", "", "访问控制权限: open, private, custom")
	executionEditCmd.Flags().StringVar(&executionWhitelist, "whitelist", "", "白名单用户分组或账号列表")
	executionEditCmd.Flags().StringVar(&executionStatus, "status", "", "执行状态: wait, doing, suspended, closed")
	executionEditCmd.Flags().StringVar(&executionDesc, "desc", "", "执行/迭代描述")
	executionEditCmd.Flags().StringVar(&executionComment, "comment", "", "操作备注说明")

	executionStartCmd.Flags().StringVar(&executionID, "id", "", "要开始的执行/迭代 ID (必填)")
	executionStartCmd.Flags().StringVar(&executionComment, "comment", "", "操作备注说明")

	executionSuspendCmd.Flags().StringVar(&executionID, "id", "", "要挂起的执行/迭代 ID (必填)")
	executionSuspendCmd.Flags().StringVar(&executionComment, "comment", "", "操作备注说明")

	executionActivateCmd.Flags().StringVar(&executionID, "id", "", "要激活的执行/迭代 ID (必填)")
	executionActivateCmd.Flags().StringVar(&executionComment, "comment", "", "操作备注说明")

	executionCloseCmd.Flags().StringVar(&executionID, "id", "", "要关闭的执行/迭代 ID (必填)")
	executionCloseCmd.Flags().StringVar(&executionComment, "comment", "", "操作备注说明")

	executionDeleteCmd.Flags().StringVar(&executionID, "id", "", "要删除的执行/迭代 ID (必填)")
	executionRestoreCmd.Flags().StringVar(&executionID, "id", "", "要恢复的执行/迭代 ID (必填)")

	executionTaskCmd.Flags().StringVar(&executionID, "id", "", "执行/迭代 ID (必填)")
	executionTaskCmd.Flags().StringVar(&executionStatus, "status", "all", "任务状态过滤: all, undone, wait, doing, done, cancel, closed")
	executionTaskCmd.Flags().StringVar(&executionOrderBy, "order-by", "id_desc", "排序字段")
	addPaginationFlags(executionTaskCmd)

	executionStoryCmd.Flags().StringVar(&executionID, "id", "", "执行/迭代 ID (必填)")
	executionStoryCmd.Flags().StringVar(&executionOrderBy, "order-by", "order_desc", "排序字段")
	addPaginationFlags(executionStoryCmd)

	executionBugCmd.Flags().StringVar(&executionID, "id", "", "执行/迭代 ID (必填)")
	executionBugCmd.Flags().StringVar(&executionOrderBy, "order-by", "id_desc", "排序字段")
	addPaginationFlags(executionBugCmd)

	executionCmd.AddCommand(executionListCmd)
	executionCmd.AddCommand(executionViewCmd)
	executionCmd.AddCommand(executionParamsCmd)
	executionCmd.AddCommand(executionCreateCmd)
	executionCmd.AddCommand(executionEditCmd)
	executionCmd.AddCommand(executionStartCmd)
	executionCmd.AddCommand(executionSuspendCmd)
	executionCmd.AddCommand(executionActivateCmd)
	executionCmd.AddCommand(executionCloseCmd)
	executionCmd.AddCommand(executionDeleteCmd)
	executionCmd.AddCommand(executionRestoreCmd)
	executionCmd.AddCommand(executionTaskCmd)
	executionCmd.AddCommand(executionStoryCmd)
	executionCmd.AddCommand(executionBugCmd)
}
