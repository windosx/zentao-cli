package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "管理禅道需求/用户故事 (Story)",
	Long:  "查询产品需求列表、查看需求详情、获取元数据、创建新需求、评审、变更、编辑或执行生命周期操作（激活、关闭、指派、删除）。",
}

var (
	storyProductID  string
	storyID         string
	storyBranch     string
	storyBrowseType string
	storyType       string
	storyOrderBy    string
	storyTitle      string
	storyPri        string
	storyEstimate   string
	storyAssignedTo string
	storyModuleID   string
	storyPlanID     string
	storySource     string
	storySourceNote string
	storyKeywords   string
	storyMailto     string
	storySpec       string
	storyVerify     string
	storyStatus     string
	storyResult     string
	storyReason     string
	storyComment    string
)

var storyListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询指定产品的需求列表（支持按分支、类型、状态与排序筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{"productID": {storyProductID}}
		if storyBranch != "" {
			params.Set("branch", storyBranch)
		}
		if storyBrowseType != "" {
			params.Set("browseType", storyBrowseType)
		}
		if storyType != "" {
			params.Set("storyType", storyType)
		}
		if storyOrderBy != "" {
			params.Set("orderBy", storyOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.StoryList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定需求的详细信息与描述",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.StoryView(ctx, storyID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取指定产品下创建需求所需的元数据字典（模块、计划、来源等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		branch := storyBranch
		if branch == "" {
			branch = "0"
		}

		data, err := client.StoryCreateParams(ctx, storyProductID, branch)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "在指定产品下创建新需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}
		if storyTitle == "" {
			return fmt.Errorf("--title 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"product": {storyProductID},
			"title":   {storyTitle},
		}
		if storyType != "" {
			params.Set("type", storyType)
		}
		if storyPri != "" {
			params.Set("pri", storyPri)
		}
		if storyEstimate != "" {
			params.Set("estimate", storyEstimate)
		}
		if storyAssignedTo != "" {
			params.Set("assignedTo", storyAssignedTo)
		}
		if storyModuleID != "" {
			params.Set("module", storyModuleID)
		}
		if storyPlanID != "" {
			params.Set("plan", storyPlanID)
		}
		if storySource != "" {
			params.Set("source", storySource)
		}
		if storySourceNote != "" {
			params.Set("sourceNote", storySourceNote)
		}
		if storyKeywords != "" {
			params.Set("keywords", storyKeywords)
		}
		if storyMailto != "" {
			params.Set("mailto", storyMailto)
		}
		if storySpec != "" {
			params.Set("spec", storySpec)
		}
		if storyVerify != "" {
			params.Set("verify", storyVerify)
		}

		data, err := client.StoryCreate(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定需求的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if storyTitle != "" {
			params.Set("title", storyTitle)
		}
		if storyType != "" {
			params.Set("type", storyType)
		}
		if storyPri != "" {
			params.Set("pri", storyPri)
		}
		if storyEstimate != "" {
			params.Set("estimate", storyEstimate)
		}
		if storyAssignedTo != "" {
			params.Set("assignedTo", storyAssignedTo)
		}
		if storyModuleID != "" {
			params.Set("module", storyModuleID)
		}
		if storyPlanID != "" {
			params.Set("plan", storyPlanID)
		}
		if storySource != "" {
			params.Set("source", storySource)
		}
		if storySourceNote != "" {
			params.Set("sourceNote", storySourceNote)
		}
		if storyKeywords != "" {
			params.Set("keywords", storyKeywords)
		}
		if storyMailto != "" {
			params.Set("mailto", storyMailto)
		}
		if storySpec != "" {
			params.Set("spec", storySpec)
		}
		if storyVerify != "" {
			params.Set("verify", storyVerify)
		}
		if storyStatus != "" {
			params.Set("status", storyStatus)
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryEdit(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "评审需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if storyResult == "" {
			return fmt.Errorf("--result 是必填参数 (pass: 确认通过, revert: 撤销变更, reject: 拒绝, clarify: 有待明确)")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"result": {storyResult},
		}
		if storyReason != "" {
			params.Set("reason", storyReason)
		}
		if storyAssignedTo != "" {
			params.Set("assignedTo", storyAssignedTo)
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryReview(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "变更需求规格与内容",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if storyTitle != "" {
			params.Set("title", storyTitle)
		}
		if storySpec != "" {
			params.Set("spec", storySpec)
		}
		if storyVerify != "" {
			params.Set("verify", storyVerify)
		}
		if storyPri != "" {
			params.Set("pri", storyPri)
		}
		if storyEstimate != "" {
			params.Set("estimate", storyEstimate)
		}
		if storyAssignedTo != "" {
			params.Set("assignedTo", storyAssignedTo)
		}
		if storyKeywords != "" {
			params.Set("keywords", storyKeywords)
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryChange(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if storyReason != "" {
			params.Set("closedReason", storyReason)
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryClose(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已关闭或已拒绝的需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if storyAssignedTo != "" {
			params.Set("assignedTo", storyAssignedTo)
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryActivate(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "指派需求给指定用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if storyAssignedTo == "" {
			return fmt.Errorf("--assigned-to 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"assignedTo": {storyAssignedTo},
		}
		if storyComment != "" {
			params.Set("comment", storyComment)
		}

		data, err := client.StoryAssign(ctx, storyID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.StoryDelete(ctx, storyID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var storyRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的需求",
	RunE: func(cmd *cobra.Command, args []string) error {
		if storyID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "story", storyID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	storyListCmd.Flags().StringVar(&storyProductID, "product", "", "所属产品 ID (必填)")
	storyListCmd.Flags().StringVar(&storyBranch, "branch", "all", "分支 ID (all 为全部/主干)")
	storyListCmd.Flags().StringVar(&storyBrowseType, "browse-type", "unclosed", "过滤条件预设: unclosed (未关闭), all (全部), draft (草稿), reviewing (评审中), changing (变更中), closed (已关闭), assignedtome (指派给我), openedbyme (由我创建), reviewedbyme (由我评审), closedbyme (由我关闭)")
	storyListCmd.Flags().StringVar(&storyType, "type", "story", "需求类型: story (用户需求/故事), requirement (软件需求)")
	storyListCmd.Flags().StringVar(&storyOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, id_asc, pri_asc, estimate_desc)")
	addPaginationFlags(storyListCmd)

	storyViewCmd.Flags().StringVar(&storyID, "id", "", "要查看的需求 ID (必填)")

	storyParamsCmd.Flags().StringVar(&storyProductID, "product", "", "所属产品 ID (必填)")
	storyParamsCmd.Flags().StringVar(&storyBranch, "branch", "all", "分支 ID (all 为全部/主干)")

	storyCreateCmd.Flags().StringVar(&storyProductID, "product", "", "所属产品 ID (必填)")
	storyCreateCmd.Flags().StringVar(&storyTitle, "title", "", "需求名称/标题 (必填)")
	storyCreateCmd.Flags().StringVar(&storyType, "type", "story", "需求类型 (story: 用户故事/需求, requirement: 软件需求)")
	storyCreateCmd.Flags().StringVar(&storyPri, "pri", "3", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	storyCreateCmd.Flags().StringVar(&storyEstimate, "estimate", "", "预计工时 (例如: 8.0)")
	storyCreateCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给的用户账号")
	storyCreateCmd.Flags().StringVar(&storyModuleID, "module", "0", "所属模块 ID")
	storyCreateCmd.Flags().StringVar(&storyPlanID, "plan", "0", "关联产品计划 ID")
	storyCreateCmd.Flags().StringVar(&storySource, "source", "", "需求来源 (customer: 客户, user: 用户, po: 产品经理, market: 市场等)")
	storyCreateCmd.Flags().StringVar(&storySourceNote, "source-note", "", "来源备注")
	storyCreateCmd.Flags().StringVar(&storyKeywords, "keywords", "", "关键词 (例如: 订单, 支付)")
	storyCreateCmd.Flags().StringVar(&storyMailto, "mailto", "", "抄送列表 (逗号分隔的账号)")
	storyCreateCmd.Flags().StringVar(&storySpec, "spec", "", "需求描述/规格说明 (作为一名<用户角色>, 我希望<实现功能>, 以便于<产生价值>)")
	storyCreateCmd.Flags().StringVar(&storyVerify, "verify", "", "验收标准")

	storyEditCmd.Flags().StringVar(&storyID, "id", "", "要修改的需求 ID (必填)")
	storyEditCmd.Flags().StringVar(&storyTitle, "title", "", "需求名称/标题")
	storyEditCmd.Flags().StringVar(&storyType, "type", "", "需求类型")
	storyEditCmd.Flags().StringVar(&storyPri, "pri", "", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	storyEditCmd.Flags().StringVar(&storyEstimate, "estimate", "", "预计工时")
	storyEditCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给的用户账号")
	storyEditCmd.Flags().StringVar(&storyModuleID, "module", "", "所属模块 ID")
	storyEditCmd.Flags().StringVar(&storyPlanID, "plan", "", "关联产品计划 ID")
	storyEditCmd.Flags().StringVar(&storySource, "source", "", "需求来源")
	storyEditCmd.Flags().StringVar(&storySourceNote, "source-note", "", "来源备注")
	storyEditCmd.Flags().StringVar(&storyKeywords, "keywords", "", "关键词")
	storyEditCmd.Flags().StringVar(&storyMailto, "mailto", "", "抄送列表")
	storyEditCmd.Flags().StringVar(&storySpec, "spec", "", "需求描述/规格说明")
	storyEditCmd.Flags().StringVar(&storyVerify, "verify", "", "验收标准")
	storyEditCmd.Flags().StringVar(&storyStatus, "status", "", "需求状态")
	storyEditCmd.Flags().StringVar(&storyComment, "comment", "", "修改备注说明")

	storyReviewCmd.Flags().StringVar(&storyID, "id", "", "要评审的需求 ID (必填)")
	storyReviewCmd.Flags().StringVar(&storyResult, "result", "pass", "评审结果: pass (确认通过), revert (撤销变更), reject (拒绝), clarify (有待明确)")
	storyReviewCmd.Flags().StringVar(&storyReason, "reason", "", "拒绝理由 (bydesign: 设计如此, duplicate: 重复, postponed: 延期, willnotdo: 不做, subversion: 下版本)")
	storyReviewCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给谁")
	storyReviewCmd.Flags().StringVar(&storyComment, "comment", "", "评审备注说明")

	storyChangeCmd.Flags().StringVar(&storyID, "id", "", "要变更的需求 ID (必填)")
	storyChangeCmd.Flags().StringVar(&storyTitle, "title", "", "变更后的需求标题")
	storyChangeCmd.Flags().StringVar(&storySpec, "spec", "", "变更后的需求描述")
	storyChangeCmd.Flags().StringVar(&storyVerify, "verify", "", "变更后的验收标准")
	storyChangeCmd.Flags().StringVar(&storyPri, "pri", "", "优先级")
	storyChangeCmd.Flags().StringVar(&storyEstimate, "estimate", "", "预计工时")
	storyChangeCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给的用户账号")
	storyChangeCmd.Flags().StringVar(&storyKeywords, "keywords", "", "关键词")
	storyChangeCmd.Flags().StringVar(&storyComment, "comment", "", "变更原因与备注")

	storyCloseCmd.Flags().StringVar(&storyID, "id", "", "要关闭的需求 ID (必填)")
	storyCloseCmd.Flags().StringVar(&storyReason, "reason", "done", "关闭原因 (done: 已完成, subversion: 延期, cancel: 取消, willnotdo: 不做)")
	storyCloseCmd.Flags().StringVar(&storyComment, "comment", "", "关闭备注说明")

	storyActivateCmd.Flags().StringVar(&storyID, "id", "", "要激活的需求 ID (必填)")
	storyActivateCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给的用户账号")
	storyActivateCmd.Flags().StringVar(&storyComment, "comment", "", "激活备注说明")

	storyAssignCmd.Flags().StringVar(&storyID, "id", "", "要指派的需求 ID (必填)")
	storyAssignCmd.Flags().StringVar(&storyAssignedTo, "assigned-to", "", "指派给的用户账号 (必填)")
	storyAssignCmd.Flags().StringVar(&storyComment, "comment", "", "指派备注说明")

	storyDeleteCmd.Flags().StringVar(&storyID, "id", "", "要删除的需求 ID (必填)")
	storyRestoreCmd.Flags().StringVar(&storyID, "id", "", "要恢复的需求 ID (必填)")

	storyCmd.AddCommand(storyListCmd)
	storyCmd.AddCommand(storyViewCmd)
	storyCmd.AddCommand(storyParamsCmd)
	storyCmd.AddCommand(storyCreateCmd)
	storyCmd.AddCommand(storyEditCmd)
	storyCmd.AddCommand(storyReviewCmd)
	storyCmd.AddCommand(storyChangeCmd)
	storyCmd.AddCommand(storyCloseCmd)
	storyCmd.AddCommand(storyActivateCmd)
	storyCmd.AddCommand(storyAssignCmd)
	storyCmd.AddCommand(storyDeleteCmd)
	storyCmd.AddCommand(storyRestoreCmd)
}
