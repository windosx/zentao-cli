package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "禅道认证与会话管理",
	Long:  "管理禅道登录状态、多 Profile 环境切换、会话有效性检查与注销。",
}

var (
	loginURL        string
	loginAccount    string
	loginPassword   string
	loginAccessMode string
	profileName     string
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录禅道并安全持久化会话凭证与环境配置",
	Long:  "使用账号和密码认证登录禅道。登录成功后，凭证与会话自动保存在 ~/.config/zentao/profiles.json 中。后续当服务端会话超时时，CLI 会自动无感重新登录续期，无需反复输入账号密码。",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL := loginURL
		if targetURL == "" {
			targetURL = runtimeOpts.URL
		}
		if targetURL == "" {
			return fmt.Errorf("--url 是必填参数（或在配置文件中指定）")
		}

		if loginAccount == "" {
			loginAccount = runtimeOpts.Account
		}
		if loginAccount == "" {
			return fmt.Errorf("--account 是必填参数")
		}

		if loginPassword == "" {
			loginPassword = runtimeOpts.Password
		}
		if loginPassword == "" {
			return fmt.Errorf("--password 是必填参数")
		}

		accessMode := loginAccessMode
		if accessMode == "" {
			accessMode = runtimeOpts.AccessMode
		}
		if accessMode == "" {
			accessMode = zentao.AccessModeGET
		}

		zcfg := zentao.Config{
			URL:        strings.TrimRight(targetURL, "/"),
			Account:    loginAccount,
			Password:   loginPassword,
			AccessMode: accessMode,
			Timeout:    client.HTTP.Timeout,
			Insecure:   runtimeOpts.Insecure,
		}

		loginClient := zentao.New(zcfg)
		ctx := context.Background()
		if err := loginClient.Login(ctx); err != nil {
			return err
		}

		// 1. Save profile and session cache with credentials for transparent auto-relogin
		_ = config.SaveProfile(config.Profile{
			Name:       profileName,
			URL:        loginClient.BaseURL,
			Account:    loginClient.Account,
			Password:   loginPassword,
			Cookie:     loginClient.Cookie,
			Rand:       loginClient.GetRand(),
			AccessMode: loginClient.AccessMode,
		})

		// 2. Persist URL and Account in global config
		globalConfigPath := config.DefaultConfigDir() + "/config.yaml"
		_ = config.SaveConfigFile(globalConfigPath, config.Options{
			URL:        loginClient.BaseURL,
			Account:    loginClient.Account,
			AccessMode: loginClient.AccessMode,
			Output:     runtimeOpts.Output,
			Timeout:    runtimeOpts.Timeout,
			Insecure:   runtimeOpts.Insecure,
		})

		// 3. Update active client
		client.BaseURL = loginClient.BaseURL
		client.Account = loginClient.Account
		client.Password = loginPassword
		client.AccessMode = loginClient.AccessMode
		client.Cookie = loginClient.Cookie
		client.SetRand(loginClient.GetRand())

		return printer.Success(map[string]any{
			"status":  "logged_in",
			"url":     loginClient.BaseURL,
			"account": loginClient.Account,
			"cookie":  loginClient.Cookie,
			"message": "登录成功，凭据与会话已持久化存储。服务端 Session 超时后将自动无感刷新，无需再次手动登录。",
		})
	},
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有已保存的环境与账号 Profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.LoadStore()
		if err != nil {
			return err
		}

		var profiles []map[string]any
		for name, p := range store.Profiles {
			active := name == store.ActiveProfile
			profiles = append(profiles, map[string]any{
				"name":       name,
				"url":        p.URL,
				"account":    p.Account,
				"accessMode": p.AccessMode,
				"active":     active,
				"updatedAt":  p.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		return printer.Success(map[string]any{
			"activeProfile": store.ActiveProfile,
			"profiles":      profiles,
		})
	},
}

var authSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "切换当前激活的 Profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileName == "" {
			return fmt.Errorf("--name 是必填参数")
		}

		p, err := config.SwitchProfile(profileName)
		if err != nil {
			return err
		}

		if client != nil {
			client.BaseURL = p.URL
			client.Account = p.Account
			client.Password = p.Password
			client.AccessMode = p.AccessMode
			client.Cookie = p.Cookie
			client.SetRand(p.Rand)
		}

		return printer.Success(map[string]any{
			"status":        "switched",
			"activeProfile": profileName,
			"url":           p.URL,
			"account":       p.Account,
		})
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前认证会话状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := ensureClientLoggedIn(ctx); err != nil {
			return err
		}

		return printer.Success(map[string]any{
			"status":      "authenticated",
			"url":         client.BaseURL,
			"account":     client.Account,
			"accessMode":  client.AccessMode,
			"hasCookie":   client.Cookie != "",
			"sessionInfo": client.Cookie,
		})
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "注销登录并清除本地会话缓存",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearSessionCache(""); err != nil {
			return err
		}
		if client != nil {
			client.Cookie = ""
			client.Password = ""
		}
		return printer.Success(map[string]any{
			"status":  "logged_out",
			"message": "本地会话缓存已成功清除",
		})
	},
}

func init() {
	authLoginCmd.Flags().StringVarP(&loginURL, "url", "u", "", "禅道服务器 Base URL")
	authLoginCmd.Flags().StringVarP(&loginAccount, "account", "a", "", "登录用户名 / 账号 (必填)")
	authLoginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "登录密码 (必填)")
	authLoginCmd.Flags().StringVarP(&loginAccessMode, "access-mode", "m", "GET", "路由模式: GET 或 PATH_INFO")
	authLoginCmd.Flags().StringVar(&profileName, "name", "", "自定义 Profile 命名 (可选)")

	authSwitchCmd.Flags().StringVar(&profileName, "name", "", "要切换的目标 Profile 名称 (必填)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authSwitchCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}
