package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "管理禅道用户",
	Long:  "查询公司用户列表或创建新用户账号（内置动态加盐 MD5 密码加密算法）。",
}

var (
	userDeptID      string
	userType        string
	userOrderBy     string
	newUsername     string
	newUserPassword string
	userRealname    string
	userRole        string
	userEmail       string
	userGender      string
	userMobile      string
	userPhone       string
	userQQ          string
)

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询用户列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if userDeptID != "" {
			params.Set("param", userDeptID)
		}
		if userType != "" {
			params.Set("type", userType)
		}
		if userOrderBy != "" {
			params.Set("orderBy", userOrderBy)
		}

		data, err := client.UserList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "创建新用户账号",
	RunE: func(cmd *cobra.Command, args []string) error {
		if newUsername == "" {
			return fmt.Errorf("--username 是必填参数")
		}
		if newUserPassword == "" {
			return fmt.Errorf("--user-password 是必填参数")
		}
		if userRealname == "" {
			return fmt.Errorf("--realname 是必填参数")
		}

		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		dept := userDeptID
		if dept == "" {
			dept = "0"
		}

		params := zentao.Params{
			"dept":      {dept},
			"account":   {newUsername},
			"password":  {newUserPassword},
			"password1": {newUserPassword},
			"password2": {newUserPassword},
			"realname":  {userRealname},
		}
		if userRole != "" {
			params.Set("role", userRole)
		}
		if userEmail != "" {
			params.Set("email", userEmail)
		}
		if userGender != "" {
			params.Set("gender", userGender)
		}
		if userMobile != "" {
			params.Set("mobile", userMobile)
		}
		if userPhone != "" {
			params.Set("phone", userPhone)
		}
		if userQQ != "" {
			params.Set("qq", userQQ)
		}

		data, err := client.UserAdd(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

func init() {
	userListCmd.Flags().StringVar(&userDeptID, "dept", "0", "部门 ID (0 代表全部部门)")
	userListCmd.Flags().StringVar(&userType, "type", "bydept", "列表类型: bydept (按部门), all (全量)")
	userListCmd.Flags().StringVar(&userOrderBy, "order-by", "id", "排序字段 (例如 id, account_asc)")

	userAddCmd.Flags().StringVar(&userDeptID, "dept", "0", "所属部门 ID")
	userAddCmd.Flags().StringVar(&newUsername, "username", "", "新用户的登录用户名 / 账号 (必填)")
	userAddCmd.Flags().StringVar(&newUserPassword, "user-password", "", "新用户的初始密码 (必填)")
	userAddCmd.Flags().StringVar(&userRealname, "realname", "", "新用户的真实姓名 (必填)")
	userAddCmd.Flags().StringVar(&userRole, "role", "dev", "用户角色 (dev, qa, pm, po 等)")
	userAddCmd.Flags().StringVar(&userEmail, "email", "", "电子邮箱地址")
	userAddCmd.Flags().StringVar(&userGender, "gender", "m", "性别: m (男), f (女)")
	userAddCmd.Flags().StringVar(&userMobile, "mobile", "", "手机号码")
	userAddCmd.Flags().StringVar(&userPhone, "phone", "", "办公电话")
	userAddCmd.Flags().StringVar(&userQQ, "qq", "", "QQ 号码")

	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userAddCmd)
}
