package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "管理禅道用户",
	Long:  "查询公司用户列表、查看用户详情、创建新用户账号、修改用户信息或删除用户账号。",
}

var (
	userID          string
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
		ctx := cmd.Context()
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
		if err := applyPagination(cmd, params); err != nil {
			return err
		}

		data, err := client.UserList(ctx, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userViewCmd = &cobra.Command{
	Use:   "view",
	Short: "查看指定用户的详细信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.UserView(ctx, userID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userParamsCmd = &cobra.Command{
	Use:   "params",
	Short: "获取创建用户所需的元数据字典（部门、用户组、角色等）",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.UserCreateParams(ctx, userDeptID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "创建新用户账号",
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

		ctx := cmd.Context()
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

var userEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "修改指定用户的信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		params := zentao.Params{}
		if userRealname != "" {
			params.Set("realname", userRealname)
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
		if userDeptID != "" {
			params.Set("dept", userDeptID)
		}

		data, err := client.UserEdit(ctx, userID, params)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "删除指定用户账号",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.UserDelete(ctx, userID)
		if err != nil {
			return err
		}
		return printer.Success(data)
	},
}

var userRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从回收站中恢复已删除的用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if userID == "" {
			return fmt.Errorf("--id 是必填参数")
		}

		ctx := cmd.Context()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		data, err := client.RestoreObject(ctx, "user", userID)
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
	addPaginationFlags(userListCmd)

	userViewCmd.Flags().StringVar(&userID, "id", "", "要查看的用户 ID (必填)")

	userParamsCmd.Flags().StringVar(&userDeptID, "dept", "0", "部门 ID (0 代表全部部门)")

	userCreateCmd.Flags().StringVar(&userDeptID, "dept", "0", "所属部门 ID")
	userCreateCmd.Flags().StringVar(&newUsername, "username", "", "新用户的登录用户名 / 账号 (必填)")
	userCreateCmd.Flags().StringVar(&newUserPassword, "user-password", "", "新用户的初始密码 (必填)")
	userCreateCmd.Flags().StringVar(&userRealname, "realname", "", "新用户的真实姓名 (必填)")
	userCreateCmd.Flags().StringVar(&userRole, "role", "dev", "用户角色 (dev, qa, pm, po 等)")
	userCreateCmd.Flags().StringVar(&userEmail, "email", "", "电子邮箱地址")
	userCreateCmd.Flags().StringVar(&userGender, "gender", "m", "性别: m (男), f (女)")
	userCreateCmd.Flags().StringVar(&userMobile, "mobile", "", "手机号码")
	userCreateCmd.Flags().StringVar(&userPhone, "phone", "", "办公电话")
	userCreateCmd.Flags().StringVar(&userQQ, "qq", "", "QQ 号码")

	userEditCmd.Flags().StringVar(&userID, "id", "", "要修改的用户 ID (必填)")
	userEditCmd.Flags().StringVar(&userRealname, "realname", "", "真实姓名")
	userEditCmd.Flags().StringVar(&userRole, "role", "", "用户角色")
	userEditCmd.Flags().StringVar(&userEmail, "email", "", "电子邮箱")
	userEditCmd.Flags().StringVar(&userGender, "gender", "", "性别 (m, f)")
	userEditCmd.Flags().StringVar(&userMobile, "mobile", "", "手机号码")
	userEditCmd.Flags().StringVar(&userPhone, "phone", "", "电话号码")
	userEditCmd.Flags().StringVar(&userQQ, "qq", "", "QQ 号码")
	userEditCmd.Flags().StringVar(&userDeptID, "dept", "", "部门 ID")

	userDeleteCmd.Flags().StringVar(&userID, "id", "", "要删除的用户 ID (必填)")
	userRestoreCmd.Flags().StringVar(&userID, "id", "", "要恢复的用户 ID (必填)")

	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userViewCmd)
	userCmd.AddCommand(userParamsCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userEditCmd)
	userCmd.AddCommand(userDeleteCmd)
	userCmd.AddCommand(userRestoreCmd)
}
