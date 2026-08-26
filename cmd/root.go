package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/internal/output"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

var (
	flagOpts config.Options

	// Global runtime variables initialized in PersistentPreRunE
	runtimeOpts *config.Options
	printer     *output.Printer
	client      *zentao.Client
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "zentao",
	Short: "禅道 (ZenTao PMS) 命令行工具（深度适配 AI Agent 与自动化脚本）",
	Long: `zentao-cli 是基于禅道官方 PHP SDK (v21.7+) 与原生 JSON 接口开发的现代化命令行工具。
提供个人工作台/待办看板（my/todo）、任务、缺陷、项目、产品等全方位能力，支持持久化会话凭据管理与确定性结构化输出。`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize config
		opts, err := config.Load(flagOpts)
		if err != nil {
			return err
		}
		runtimeOpts = opts
		printer = output.New(opts.Output)
		printer.Out = cmd.OutOrStdout()
		printer.Err = cmd.ErrOrStderr()

		// Create ZenTao Client with active configuration
		zcfg := opts.ToZentaoConfig()
		client = zentao.New(zcfg)
		client.OnSessionRefreshed = func(cookie, rand string) {
			config.UpdateActiveProfileCookie(cookie, rand)
		}

		// Auto-bind active profile credentials and session
		if profile, err := config.GetActiveProfile(""); err == nil && profile != nil {
			if client.BaseURL == "" {
				client.BaseURL = strings.TrimRight(profile.URL, "/")
			}
			if client.Account == "" {
				client.Account = profile.Account
			}
			if client.Password == "" && profile.Password != "" {
				client.Password = profile.Password
			}
			if client.GetRand() == "" {
				client.SetRand(profile.Rand)
			}
			if profile.Cookie != "" {
				client.Cookie = profile.Cookie
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		code, category := classifyError(err)
		if printer != nil {
			printer.Fail(code, category, err.Error(), nil)
		} else {
			fmt.Fprintf(os.Stderr, "错误 [%s]: %v\n", category, err)
		}
		os.Exit(code)
	}
}

func classifyError(err error) (int, string) {
	if err == nil {
		return output.ExitCodeSuccess, "none"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not logged in") || strings.Contains(msg, "login") || strings.Contains(msg, "auth") || strings.Contains(msg, "session") || strings.Contains(msg, "超时") || strings.Contains(msg, "登入"):
		return output.ExitCodeAuth, "auth"
	case strings.Contains(msg, "required") || strings.Contains(msg, "invalid") || strings.Contains(msg, "flag") || strings.Contains(msg, "unknown"):
		return output.ExitCodeValidation, "validation"
	default:
		return output.ExitCodeAPI, "api"
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&flagOpts.ConfigFile, "config", "c", "", "配置文件路径 (默认: .zentao.yaml 或 ~/.config/zentao/config.yaml)")
	RootCmd.PersistentFlags().StringVarP(&flagOpts.URL, "url", "u", "", "禅道服务器 Base URL 地址 (环境变量: ZENTAO_URL)")
	RootCmd.PersistentFlags().StringVarP(&flagOpts.AccessMode, "access-mode", "m", "GET", "接口路由模式: GET 或 PATH_INFO (环境变量: ZENTAO_ACCESS_MODE)")
	RootCmd.PersistentFlags().BoolVarP(&flagOpts.Insecure, "insecure", "k", false, "允许非安全的 HTTPS 连接 (忽略 SSL 证书校验) (环境变量: ZENTAO_INSECURE)")
	RootCmd.PersistentFlags().StringVarP(&flagOpts.Output, "output", "o", "json", "输出格式: json, raw-json, yaml, table, text (环境变量: ZENTAO_OUTPUT)")
	RootCmd.PersistentFlags().StringVar(&flagOpts.Timeout, "timeout", "30s", "HTTP 请求超时时长 (环境变量: ZENTAO_TIMEOUT)")

	// Register subcommands
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(myCmd)
	RootCmd.AddCommand(todoCmd)
	RootCmd.AddCommand(taskCmd)
	RootCmd.AddCommand(bugCmd)
	RootCmd.AddCommand(projectCmd)
	RootCmd.AddCommand(productCmd)
	RootCmd.AddCommand(userCmd)
	RootCmd.AddCommand(deptCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(schemaCmd)
	RootCmd.AddCommand(skillCmd)
	RootCmd.AddCommand(versionCmd)
}

// ensureClientLoggedIn verifies that the client has an active authenticated session or credentials to auto-login.
func ensureClientLoggedIn(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("客户端未初始化")
	}

	// 1. If client already has a session cookie, try using it (if expired, callRoute will transparently re-login and retry)
	if client.Cookie != "" {
		return nil
	}

	// 2. Try loading active profile
	if profile, err := config.GetActiveProfile(""); err == nil && profile != nil {
		if client.BaseURL == "" {
			client.BaseURL = strings.TrimRight(profile.URL, "/")
		}
		if client.Account == "" {
			client.Account = profile.Account
		}
		if client.Password == "" && profile.Password != "" {
			client.Password = profile.Password
		}
		if client.GetRand() == "" {
			client.SetRand(profile.Rand)
		}
		if profile.Cookie != "" {
			client.Cookie = profile.Cookie
			return nil
		}
	}

	// 3. If no cookie but credentials exist, perform transparent login
	if client.BaseURL != "" && client.Account != "" && client.Password != "" {
		return client.Login(ctx)
	}

	return fmt.Errorf("尚未登录，请先执行: zentao auth login --url <url> --account <account> --password <password>")
}
