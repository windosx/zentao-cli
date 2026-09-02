package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/internal/output"
	"github.com/windosx/zentao-cli/internal/secret"
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

		// Auto-bind active profile credentials and session
		if profile, err := config.GetActiveProfile(""); err == nil && profile != nil {
			bindProfileToClient(profile)
		}

		// Load password from keyring if not provided in flags/profile
		if client.Password == "" && client.BaseURL != "" && client.Account != "" {
			if pw, err := secret.Get(client.BaseURL, client.Account); err == nil {
				client.Password = pw
			}
		}

		client.OnSessionRefreshed = func(cookie, rand string) {
			config.UpdateActiveProfileCookie(cookie, rand)
		}

		// Silently auto-sync existing installed SKILL.md to the latest version on binary update
		AutoSyncInstalledSkills()

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

// classifyError maps an error to an exit code and category. Structured
// zentao errors are classified by kind; only cobra flag/command usage errors
// fall back to narrow string matching.
func classifyError(err error) (int, string) {
	if err == nil {
		return output.ExitCodeSuccess, "none"
	}
	if zentao.IsAuthError(err) {
		return output.ExitCodeAuth, "auth"
	}
	if errors.Is(err, zentao.ErrValidation) {
		return output.ExitCodeValidation, "validation"
	}
	msg := strings.ToLower(err.Error())
	// Cobra usage errors: unknown command/flag, missing required flag.
	if strings.Contains(msg, "unknown command") ||
		strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "unknown shorthand flag") ||
		strings.Contains(msg, "required flag") ||
		strings.Contains(msg, "flag needs an argument") ||
		strings.Contains(msg, "invalid argument") {
		return output.ExitCodeValidation, "validation"
	}
	return output.ExitCodeAPI, "api"
}

// maskCookie redacts a session cookie for display: "name=****abcd".
// The full value is only shown with explicit --show-secrets.
func maskCookie(cookie string) string {
	if cookie == "" {
		return ""
	}
	name, value, ok := strings.Cut(cookie, "=")
	if !ok {
		return "****" + cookie[len(cookie)-4:]
	}
	if len(value) <= 4 {
		return name + "=****" + value
	}
	return name + "=****" + value[len(value)-4:]
}

// bindProfileToClient copies credentials and session state from a saved
// profile into the global client, filling only empty fields.
func bindProfileToClient(p *config.Profile) {
	if client == nil || p == nil {
		return
	}
	if client.BaseURL == "" {
		client.BaseURL = strings.TrimRight(p.URL, "/")
	}
	if client.Account == "" {
		client.Account = p.Account
	}
	if client.Password == "" && p.Password != "" {
		client.Password = p.Password
	}
	if client.GetRand() == "" {
		client.SetRand(p.Rand)
	}
	if client.Cookie == "" {
		client.Cookie = p.Cookie
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
	RootCmd.AddCommand(storyCmd)
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

	// 1. Try loading active profile if client fields are empty
	if client.BaseURL == "" || client.Account == "" || client.Cookie == "" {
		if profile, err := config.GetActiveProfile(""); err == nil && profile != nil {
			bindProfileToClient(profile)
		}
	}

	// 2. Load password from keyring if needed
	if client.Password == "" && client.BaseURL != "" && client.Account != "" {
		if pw, err := secret.Get(client.BaseURL, client.Account); err == nil {
			client.Password = pw
		}
	}

	// 3. If client already has a session cookie, try using it (if expired, callRoute will transparently re-login and retry)
	if client.Cookie != "" {
		return nil
	}

	// 4. If no cookie but credentials exist, perform transparent login
	if client.BaseURL != "" && client.Account != "" && client.Password != "" {
		return client.Login(ctx)
	}

	return fmt.Errorf("尚未登录，请先执行: zentao auth login --url <url> --account <account> --password <password>")
}
