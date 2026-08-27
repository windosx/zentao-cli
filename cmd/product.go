package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var productCmd = &cobra.Command{
	Use:   "product",
	Short: "管理禅道产品",
	Long:  "查询禅道产品列表或创建新产品，支持按产品状态（未关闭、正常、已关闭、全部）、所属项目集等条件筛选与排序。",
}

var (
	productStatus    string
	productOrderBy   string
	productLineID    string
	productProgramID string
	productName      string
	productCode      string
	productPO        string
	productQD        string
	productRD        string
	productACL       string
	productDesc      string
)

var productListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询产品列表（支持按状态、产品线、项目集筛选）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
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

		data, err := client.ProductList(ctx, params)
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
		ctx := context.Background()
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

var productAddCmd = &cobra.Command{
	Use:   "add",
	Short: "创建新产品",
	RunE: func(cmd *cobra.Command, args []string) error {
		if productName == "" {
			return fmt.Errorf("--name 是必填参数")
		}
		if productCode == "" {
			return fmt.Errorf("--code 是必填参数")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{
			"name": {productName},
			"code": {productCode},
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

func init() {
	productListCmd.Flags().StringVar(&productStatus, "status", "noclosed", "产品状态过滤: noclosed (未关闭/正常), all (全部产品), normal (正常运营中), closed (已关闭)")
	productListCmd.Flags().StringVar(&productLineID, "line", "0", "按产品线 ID 过滤 (0 代表全量)")
	productListCmd.Flags().StringVar(&productProgramID, "program", "0", "按所属项目集/项目 ID 过滤 (0 代表全量/无项目集)")
	productListCmd.Flags().StringVar(&productOrderBy, "order-by", "order_desc", "排序字段 (例如: order_desc, order_asc, id_desc, id_asc, name_asc)")

	productParamsCmd.Flags().StringVar(&productProgramID, "program", "0", "按所属项目集/项目 ID 过滤 (0 代表全量/无项目集)")

	productAddCmd.Flags().StringVar(&productName, "name", "", "产品名称 (必填)")
	productAddCmd.Flags().StringVar(&productCode, "code", "", "产品代号 / 英文缩写 (必填)")
	productAddCmd.Flags().StringVar(&productPO, "po", "", "产品负责人 (PO) 账号")
	productAddCmd.Flags().StringVar(&productQD, "qd", "", "测试负责人 (QD) 账号")
	productAddCmd.Flags().StringVar(&productRD, "rd", "", "发布负责人 (RD) 账号")
	productAddCmd.Flags().StringVar(&productACL, "acl", "open", "访问权限控制: open (公开), custom (自定义白名单), private (私有)")
	productAddCmd.Flags().StringVar(&productDesc, "desc", "", "产品描述说明")

	productCmd.AddCommand(productListCmd)
	productCmd.AddCommand(productParamsCmd)
	productCmd.AddCommand(productAddCmd)
}
