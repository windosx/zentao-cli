package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var productCmd = &cobra.Command{
	Use:   "product",
	Short: "管理禅道产品",
	Long:  "查询禅道产品列表、查看产品详情、创建新产品、编辑产品信息或执行生命周期操作（关闭、激活、删除）。",
}

var (
	productID        string
	productStatus    string
	productOrderBy   string
	productLineID    string
	productProgramID string
	productName      string
	productCode      string
	productType      string
	productPO        string
	productQD        string
	productRD        string
	productACL       string
	productWhitelist string
	productDesc      string
	productComment   string
)

var productListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询产品列表（支持按状态、产品线、项目集筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if productStatus != "" {
			params.Set("browseType", productStatus)
			params.Set("status", productStatus)
		}
		if productOrderBy != "" {
			params.Set("orderBy", productOrderBy)
		}
		if productLineID != "" {
			params.Set("param", productLineID)
		}
		if productProgramID != "" {
			params.Set("programID", productProgramID)
		}
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.ProductList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定产品的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProductView(ctx, productID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取创建产品所需的元数据字典（产品线、PO/QD/RD 负责人、用户组等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProductCreateParams(ctx, productProgramID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "创建新产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productName == "" {
			return fmt.Errorf("--name 是必填参数")
		}
		if productCode == "" {
			return fmt.Errorf("--code 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {productName},
			"code": {productCode},
		}
		if productLineID != "" {
			params.Set("line", productLineID)
		}
		if productProgramID != "" {
			params.Set("program", productProgramID)
		}
		if productType != "" {
			params.Set("type", productType)
		}
		if productPO != "" {
			params.Set("PO", productPO)
		}
		if productQD != "" {
			params.Set("QD", productQD)
		}
		if productRD != "" {
			params.Set("RD", productRD)
		}
		if productACL != "" {
			params.Set("acl", productACL)
		}
		if productWhitelist != "" {
			params.Set("whitelist", productWhitelist)
		}
		if productStatus != "" {
			params.Set("status", productStatus)
		}
		if productDesc != "" {
			params.Set("desc", productDesc)
		}

		data, err := client.ProductAdd(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定产品的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if productName != "" {
			params.Set("name", productName)
		}
		if productCode != "" {
			params.Set("code", productCode)
		}
		if productLineID != "" {
			params.Set("line", productLineID)
		}
		if productProgramID != "" {
			params.Set("program", productProgramID)
		}
		if productType != "" {
			params.Set("type", productType)
		}
		if productPO != "" {
			params.Set("PO", productPO)
		}
		if productQD != "" {
			params.Set("QD", productQD)
		}
		if productRD != "" {
			params.Set("RD", productRD)
		}
		if productACL != "" {
			params.Set("acl", productACL)
		}
		if productWhitelist != "" {
			params.Set("whitelist", productWhitelist)
		}
		if productStatus != "" {
			params.Set("status", productStatus)
		}
		if productDesc != "" {
			params.Set("desc", productDesc)
		}
		if productComment != "" {
			params.Set("comment", productComment)
		}

		data, err := client.ProductEdit(ctx, productID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "关闭产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if productComment != "" {
			params.Set("comment", productComment)
		}

		data, err := client.ProductClose(ctx, productID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "激活已关闭的产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if productComment != "" {
			params.Set("comment", productComment)
		}

		data, err := client.ProductActivate(ctx, productID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.ProductDelete(ctx, productID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var productRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "product", productID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	productListCmd.Flags().StringVar(&productStatus, "status", "noclosed", "产品状态过滤: noclosed (未关闭/正常), all (全部产品), normal (正常运营中), closed (已关闭)")
	productListCmd.Flags().StringVar(&productLineID, "line", "0", "按产品线 ID 过滤 (0 代表全量)")
	productListCmd.Flags().StringVar(&productProgramID, "program", "0", "按所属项目集/项目 ID 过滤 (0 代表全量/无项目集)")
	productListCmd.Flags().StringVar(&productOrderBy, "order-by", "order_desc", "排序字段 (例如: order_desc, order_asc, id_desc, id_asc, name_asc)")
	addPaginationFlags(productListCmd)

	productViewCmd.Flags().StringVar(&productID, "id", "", "要查看的产品 ID (必填)")

	productParamsCmd.Flags().StringVar(&productProgramID, "program", "0", "按所属项目集/项目 ID 过滤 (0 代表全量/无项目集)")

	productCreateCmd.Flags().StringVar(&productName, "name", "", "产品名称 (必填)")
	productCreateCmd.Flags().StringVar(&productCode, "code", "", "产品代号 / 英文缩写 (必填)")
	productCreateCmd.Flags().StringVar(&productLineID, "line", "0", "所属产品线 ID")
	productCreateCmd.Flags().StringVar(&productProgramID, "program", "0", "所属项目集 ID")
	productCreateCmd.Flags().StringVar(&productType, "type", "normal", "产品类型 (normal: 正常, branch: 多分支, platform: 多平台)")
	productCreateCmd.Flags().StringVar(&productPO, "po", "", "产品负责人 (PO) 账号")
	productCreateCmd.Flags().StringVar(&productQD, "qd", "", "测试负责人 (QD) 账号")
	productCreateCmd.Flags().StringVar(&productRD, "rd", "", "发布负责人 (RD) 账号")
	productCreateCmd.Flags().StringVar(&productACL, "acl", "open", "访问权限控制: open (公开), custom (自定义白名单), private (私有)")
	productCreateCmd.Flags().StringVar(&productWhitelist, "whitelist", "", "白名单列表 (逗号分隔)")
	productCreateCmd.Flags().StringVar(&productStatus, "status", "normal", "产品初始状态 (normal, closed)")
	productCreateCmd.Flags().StringVar(&productDesc, "desc", "", "产品描述说明")

	productEditCmd.Flags().StringVar(&productID, "id", "", "要修改的产品 ID (必填)")
	productEditCmd.Flags().StringVar(&productName, "name", "", "产品名称")
	productEditCmd.Flags().StringVar(&productCode, "code", "", "产品代号")
	productEditCmd.Flags().StringVar(&productLineID, "line", "", "产品线 ID")
	productEditCmd.Flags().StringVar(&productProgramID, "program", "", "所属项目集 ID")
	productEditCmd.Flags().StringVar(&productType, "type", "", "产品类型")
	productEditCmd.Flags().StringVar(&productPO, "po", "", "产品负责人 (PO) 账号")
	productEditCmd.Flags().StringVar(&productQD, "qd", "", "测试负责人 (QD) 账号")
	productEditCmd.Flags().StringVar(&productRD, "rd", "", "发布负责人 (RD) 账号")
	productEditCmd.Flags().StringVar(&productACL, "acl", "", "访问权限控制")
	productEditCmd.Flags().StringVar(&productWhitelist, "whitelist", "", "白名单列表")
	productEditCmd.Flags().StringVar(&productStatus, "status", "", "产品状态")
	productEditCmd.Flags().StringVar(&productDesc, "desc", "", "产品描述说明")
	productEditCmd.Flags().StringVar(&productComment, "comment", "", "修改备注说明")

	productCloseCmd.Flags().StringVar(&productID, "id", "", "要关闭的产品 ID (必填)")
	productCloseCmd.Flags().StringVar(&productComment, "comment", "", "关闭备注说明")

	productActivateCmd.Flags().StringVar(&productID, "id", "", "要激活的产品 ID (必填)")
	productActivateCmd.Flags().StringVar(&productComment, "comment", "", "激活备注说明")

	productDeleteCmd.Flags().StringVar(&productID, "id", "", "要删除的产品 ID (必填)")
	productRestoreCmd.Flags().StringVar(&productID, "id", "", "要恢复的产品 ID (必填)")

	productCmd.AddCommand(productListCmd)
	productCmd.AddCommand(productViewCmd)
	productCmd.AddCommand(productParamsCmd)
	productCmd.AddCommand(productCreateCmd)
	productCmd.AddCommand(productEditCmd)
	productCmd.AddCommand(productCloseCmd)
	productCmd.AddCommand(productActivateCmd)
	productCmd.AddCommand(productDeleteCmd)
	productCmd.AddCommand(productRestoreCmd)
}
