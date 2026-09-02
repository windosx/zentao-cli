package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var bugCmd = &cobra.Command{
	Use:   "bug",
	Short: "管理禅道缺陷 (Bug)",
	Long:  "查询产品 Bug 列表、查看 Bug 详情、获取元数据、提交新 Bug、编辑或执行生命周期操作（解决、关闭、激活、确认、指派、删除）。",
}

var (
	bugProductID     string
	bugID            string
	bugBranch        string
	bugBrowseType    string
	bugOrderBy       string
	bugTitle         string
	bugSeverity      string
	bugPri           string
	bugType          string
	bugAssignedTo    string
	bugSteps         string
	bugOpenedBuild   string
	bugProjectID     string
	bugStoryID       string
	bugModuleID      string
	bugKeywords      string
	bugMailto        string
	bugOS            string
	bugBrowser       string
	bugHardware      string
	bugFound         string
	bugDeadline      string
	bugDuplicateBug  string
	bugStatus        string
	bugResolution    string
	bugResolvedBuild string
	bugResolvedDate  string
	bugComment       string
)

var bugListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询指定产品的 Bug 列表（支持全量预置过滤与排序）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{"productID": {bugProductID}}
		if bugBranch != "" {
			params.Set("branch", bugBranch)
		}
		if bugBrowseType != "" {
			params.Set("browseType", bugBrowseType)
		}
		if bugOrderBy != "" {
			params.Set("orderBy", bugOrderBy)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.BugList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定 Bug 的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.BugView(ctx, bugID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取指定产品下创建 Bug 所需的元数据与字典（类型、版本、指派人等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		branch := bugBranch
		if branch == "" {
			branch = "0"
		}

		data, err := client.BugCreateParams(ctx, bugProductID, branch)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "在指定产品下提交新 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugProductID == "" {
			return fmt.Errorf("--product 是必填参数")
		}
		if bugTitle == "" {
			return fmt.Errorf("--title 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		build := bugOpenedBuild
		if build == "" {
			build = "trunk"
		}

		params := zentao.Params{
			"product":        {bugProductID},
			"title":          {bugTitle},
			"openedBuild":    {build},
			"openedBuild[]":  {build},
			"openedBuild[0]": {build},
		}
		if bugSeverity != "" {
			params.Set("severity", bugSeverity)
		}
		if bugPri != "" {
			params.Set("pri", bugPri)
		}
		if bugType != "" {
			params.Set("type", bugType)
		}
		if bugAssignedTo != "" {
			params.Set("assignedTo", bugAssignedTo)
		}
		if bugSteps != "" {
			params.Set("steps", bugSteps)
		}
		if bugProjectID != "" {
			params.Set("project", bugProjectID)
		}
		if bugStoryID != "" {
			params.Set("story", bugStoryID)
		}
		if bugModuleID != "" {
			params.Set("module", bugModuleID)
		}
		if bugKeywords != "" {
			params.Set("keywords", bugKeywords)
		}
		if bugMailto != "" {
			params.Set("mailto", bugMailto)
		}
		if bugOS != "" {
			params.Set("os", bugOS)
		}
		if bugBrowser != "" {
			params.Set("browser", bugBrowser)
		}
		if bugHardware != "" {
			params.Set("hardware", bugHardware)
		}
		if bugFound != "" {
			params.Set("found", bugFound)
		}
		if bugDeadline != "" {
			params.Set("deadline", bugDeadline)
		}

		data, err := client.BugCreate(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定 Bug 的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if bugTitle != "" {
			params.Set("title", bugTitle)
		}
		if bugSeverity != "" {
			params.Set("severity", bugSeverity)
		}
		if bugPri != "" {
			params.Set("pri", bugPri)
		}
		if bugType != "" {
			params.Set("type", bugType)
		}
		if bugAssignedTo != "" {
			params.Set("assignedTo", bugAssignedTo)
		}
		if bugSteps != "" {
			params.Set("steps", bugSteps)
		}
		if bugOpenedBuild != "" {
			params.Set("openedBuild", bugOpenedBuild)
			params["openedBuild[]"] = []string{bugOpenedBuild}
			params["openedBuild[0]"] = []string{bugOpenedBuild}
		}
		if bugProjectID != "" {
			params.Set("project", bugProjectID)
		}
		if bugStoryID != "" {
			params.Set("story", bugStoryID)
		}
		if bugModuleID != "" {
			params.Set("module", bugModuleID)
		}
		if bugKeywords != "" {
			params.Set("keywords", bugKeywords)
		}
		if bugMailto != "" {
			params.Set("mailto", bugMailto)
		}
		if bugOS != "" {
			params.Set("os", bugOS)
		}
		if bugBrowser != "" {
			params.Set("browser", bugBrowser)
		}
		if bugHardware != "" {
			params.Set("hardware", bugHardware)
		}
		if bugFound != "" {
			params.Set("found", bugFound)
		}
		if bugDeadline != "" {
			params.Set("deadline", bugDeadline)
		}
		if bugStatus != "" {
			params.Set("status", bugStatus)
		}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugEdit(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugResolveParamsCmd = &cobra.Command{
	Use:   "resolve-params",
	Short: "获取解决 Bug 所需的元数据字典（解决方案列表、构建版本、当前 Bug 详情等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.BugResolveParams(ctx, bugID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "解决 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if bugResolution == "" {
			return fmt.Errorf("--resolution 是必填参数 (bydesign: 设计如此, duplicate: 重复Bug, external: 外部原因, fixed: 已解决, notrepro: 无法重现, postponed: 延期处理, willnotfix: 不予解决, tostory: 转为需求)")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		build := bugResolvedBuild
		if build == "" {
			build = "trunk"
		}

		params := zentao.Params{
			"resolution":    {bugResolution},
			"resolvedBuild": {build},
		}
		if bugDuplicateBug != "" {
			params.Set("duplicateBug", bugDuplicateBug)
		}
		if bugResolvedDate != "" {
			params.Set("resolvedDate", bugResolvedDate)
		}
		if bugAssignedTo != "" {
			params.Set("assignedTo", bugAssignedTo)
		}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugResolve(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭已解决或无需处理的 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugClose(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已解决或已关闭的 Bug (重新打开)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if bugOpenedBuild != "" {
			params.Set("openedBuild", bugOpenedBuild)
		}
		if bugAssignedTo != "" {
			params.Set("assignedTo", bugAssignedTo)
		}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugActivate(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "指派 Bug 给指定用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}
		if bugAssignedTo == "" {
			return fmt.Errorf("--assigned-to 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"assignedTo": {bugAssignedTo},
		}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugAssign(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugConfirmCmd = &cobra.Command{
	Use:   "confirm",
	Short: "确认 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if bugPri != "" {
			params.Set("pri", bugPri)
		}
		if bugAssignedTo != "" {
			params.Set("assignedTo", bugAssignedTo)
		}
		if bugComment != "" {
			params.Set("comment", bugComment)
		}

		data, err := client.BugConfirm(ctx, bugID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.BugDelete(ctx, bugID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var bugRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的 Bug",
	RunE: func(cmd *cobra.Command, args []string) error {
		if bugID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "bug", bugID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	bugListCmd.Flags().StringVar(&bugProductID, "product", "", "所属产品 ID (必填)")
	bugListCmd.Flags().StringVar(&bugBranch, "branch", "all", "分支 ID (all 为全部/主干)")
	bugListCmd.Flags().StringVar(&bugBrowseType, "browse-type", "all", "过滤条件预设: all (全部), unclosed (未关闭), assigntome (指派给我), openedbyme (由我创建), resolvedbyme (由我解决), closedbyme (由我关闭), needconfirm (待确认), toclosed (待关闭), unconfirmed (未确认), unassigned (未指派), longlife (久未解决), postponed (被延期), overdue (已过期)")
	bugListCmd.Flags().StringVar(&bugOrderBy, "order-by", "id_desc", "排序字段 (例如: id_desc, id_asc, pri_asc, severity_desc, openedDate_desc)")
	addPaginationFlags(bugListCmd)

	bugViewCmd.Flags().StringVar(&bugID, "id", "", "要查看的 Bug ID (必填)")

	bugParamsCmd.Flags().StringVar(&bugProductID, "product", "", "所属产品 ID (必填)")
	bugParamsCmd.Flags().StringVar(&bugBranch, "branch", "all", "分支 ID (all 为全部/主干)")

	bugCreateCmd.Flags().StringVar(&bugProductID, "product", "", "所属产品 ID (必填)")
	bugCreateCmd.Flags().StringVar(&bugTitle, "title", "", "Bug 标题 (必填)")
	bugCreateCmd.Flags().StringVar(&bugSeverity, "severity", "3", "严重程度 (1=致命, 2=严重, 3=一般, 4=轻微)")
	bugCreateCmd.Flags().StringVar(&bugPri, "pri", "3", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	bugCreateCmd.Flags().StringVar(&bugType, "type", "codeerror", "Bug 类型 (codeerror: 代码错误, interface: 界面优化, designchange: 设计变更, newfeature: 新需求, designdefect: 设计缺陷, trackthings: 跟踪事务, badquality: 质量差, other: 其他)")
	bugCreateCmd.Flags().StringVar(&bugOpenedBuild, "opened-build", "trunk", "影响版本 / 构建版本 (例如: trunk, 1.0.0)")
	bugCreateCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "指派给的用户账号")
	bugCreateCmd.Flags().StringVar(&bugSteps, "steps", "", "重现步骤 (支持富文本/换行)")
	bugCreateCmd.Flags().StringVar(&bugProjectID, "project", "0", "关联所属项目 / 执行 ID")
	bugCreateCmd.Flags().StringVar(&bugStoryID, "story", "0", "关联所属需求/故事 ID")
	bugCreateCmd.Flags().StringVar(&bugModuleID, "module", "0", "所属模块 ID")
	bugCreateCmd.Flags().StringVar(&bugKeywords, "keywords", "", "Bug 关键词 (例如: 崩溃, 权限)")
	bugCreateCmd.Flags().StringVar(&bugMailto, "mailto", "", "抄送列表 (逗号分隔的账号)")
	bugCreateCmd.Flags().StringVar(&bugOS, "os", "", "操作系统 (windows, mac, linux, android, ios, all等)")
	bugCreateCmd.Flags().StringVar(&bugBrowser, "browser", "", "浏览器 (chrome, firefox, safari, edge, all等)")
	bugCreateCmd.Flags().StringVar(&bugHardware, "hardware", "", "硬件平台")
	bugCreateCmd.Flags().StringVar(&bugFound, "found", "", "发现于版本")
	bugCreateCmd.Flags().StringVar(&bugDeadline, "deadline", "", "截止解决日期 (格式: YYYY-MM-DD)")

	bugEditCmd.Flags().StringVar(&bugID, "id", "", "要修改的 Bug ID (必填)")
	bugEditCmd.Flags().StringVar(&bugTitle, "title", "", "Bug 标题")
	bugEditCmd.Flags().StringVar(&bugSeverity, "severity", "", "严重程度 (1=致命, 2=严重, 3=一般, 4=轻微)")
	bugEditCmd.Flags().StringVar(&bugPri, "pri", "", "优先级 (1=最高, 2=高, 3=中, 4=低)")
	bugEditCmd.Flags().StringVar(&bugType, "type", "", "Bug 类型")
	bugEditCmd.Flags().StringVar(&bugOpenedBuild, "opened-build", "", "影响版本")
	bugEditCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "指派给的用户账号")
	bugEditCmd.Flags().StringVar(&bugSteps, "steps", "", "重现步骤")
	bugEditCmd.Flags().StringVar(&bugProjectID, "project", "", "关联项目 ID")
	bugEditCmd.Flags().StringVar(&bugStoryID, "story", "", "关联需求 ID")
	bugEditCmd.Flags().StringVar(&bugModuleID, "module", "", "所属模块 ID")
	bugEditCmd.Flags().StringVar(&bugKeywords, "keywords", "", "关键词")
	bugEditCmd.Flags().StringVar(&bugMailto, "mailto", "", "抄送列表")
	bugEditCmd.Flags().StringVar(&bugOS, "os", "", "操作系统")
	bugEditCmd.Flags().StringVar(&bugBrowser, "browser", "", "浏览器")
	bugEditCmd.Flags().StringVar(&bugHardware, "hardware", "", "硬件平台")
	bugEditCmd.Flags().StringVar(&bugFound, "found", "", "发现于版本")
	bugEditCmd.Flags().StringVar(&bugDeadline, "deadline", "", "截止日期")
	bugEditCmd.Flags().StringVar(&bugStatus, "status", "", "Bug 状态")
	bugEditCmd.Flags().StringVar(&bugComment, "comment", "", "修改备注说明")

	bugResolveCmd.Flags().StringVar(&bugID, "id", "", "要解决的 Bug ID (必填)")
	bugResolveCmd.Flags().StringVar(&bugResolution, "resolution", "fixed", "解决方案 (fixed: 已解决, bydesign: 设计如此, duplicate: 重复Bug, external: 外部原因, notrepro: 无法重现, postponed: 延期处理, willnotfix: 不予解决, tostory: 转为需求)")
	bugResolveCmd.Flags().StringVar(&bugResolvedBuild, "resolved-build", "trunk", "解决该 Bug 的构建版本号 (例如: trunk, 1.0.1)")
	bugResolveCmd.Flags().StringVar(&bugDuplicateBug, "duplicate-bug", "", "重复的 Bug ID (当 resolution 为 duplicate 时提供)")
	bugResolveCmd.Flags().StringVar(&bugResolvedDate, "resolved-date", "", "解决日期 (格式: YYYY-MM-DD)")
	bugResolveCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "解决后指派给 (默认给 Bug 创建者进行关闭/确认)")
	bugResolveCmd.Flags().StringVar(&bugComment, "comment", "", "解决方案备注说明")

	bugResolveParamsCmd.Flags().StringVar(&bugID, "id", "", "要解决的 Bug ID (必填)")

	bugCloseCmd.Flags().StringVar(&bugID, "id", "", "要关闭的 Bug ID (必填)")
	bugCloseCmd.Flags().StringVar(&bugComment, "comment", "", "关闭备注说明")

	bugActivateCmd.Flags().StringVar(&bugID, "id", "", "要激活的 Bug ID (必填)")
	bugActivateCmd.Flags().StringVar(&bugOpenedBuild, "opened-build", "trunk", "重新激活影响的构建版本")
	bugActivateCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "指派给的用户账号")
	bugActivateCmd.Flags().StringVar(&bugComment, "comment", "", "激活备注说明")

	bugAssignCmd.Flags().StringVar(&bugID, "id", "", "要指派的 Bug ID (必填)")
	bugAssignCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "指派给的用户账号 (必填)")
	bugAssignCmd.Flags().StringVar(&bugComment, "comment", "", "指派备注说明")

	bugConfirmCmd.Flags().StringVar(&bugID, "id", "", "要确认的 Bug ID (必填)")
	bugConfirmCmd.Flags().StringVar(&bugPri, "pri", "", "优先级")
	bugConfirmCmd.Flags().StringVar(&bugAssignedTo, "assigned-to", "", "指派给的用户账号")
	bugConfirmCmd.Flags().StringVar(&bugComment, "comment", "", "确认备注说明")

	bugDeleteCmd.Flags().StringVar(&bugID, "id", "", "要删除的 Bug ID (必填)")
	bugRestoreCmd.Flags().StringVar(&bugID, "id", "", "要恢复的 Bug ID (必填)")

	bugCmd.AddCommand(bugListCmd)
	bugCmd.AddCommand(bugViewCmd)
	bugCmd.AddCommand(bugParamsCmd)
	bugCmd.AddCommand(bugCreateCmd)
	bugCmd.AddCommand(bugEditCmd)
	bugCmd.AddCommand(bugResolveParamsCmd)
	bugCmd.AddCommand(bugResolveCmd)
	bugCmd.AddCommand(bugCloseCmd)
	bugCmd.AddCommand(bugActivateCmd)
	bugCmd.AddCommand(bugAssignCmd)
	bugCmd.AddCommand(bugConfirmCmd)
	bugCmd.AddCommand(bugDeleteCmd)
	bugCmd.AddCommand(bugRestoreCmd)
}
