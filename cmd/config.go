package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理 zentao-cli 配置文件",
	Long:  "初始化或查看 zentao-cli 配置文件及生效配置参数。",
}

var (
	configInitPath   string
	configInitGlobal bool
	configInitForce  bool
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化默认配置文件模板",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := configInitPath
		if targetPath == "" {
			if configInitGlobal {
				targetPath = filepath.Join(config.DefaultConfigDir(), "config.yaml")
			} else {
				targetPath = ".zentao.yaml"
			}
		}

		if _, err := os.Stat(targetPath); err == nil && !configInitForce {
			return fmt.Errorf("配置文件 %q 已存在；请使用 --force 覆盖", targetPath)
		}

		sampleOpts := config.Options{
			URL:        "https://zentao.example.com",
			Account:    "admin",
			Password:   "your-password",
			AccessMode: "GET",
			Output:     "json",
			Timeout:    "30s",
			Insecure:   false,
		}

		absPath, _ := filepath.Abs(targetPath)
		if err := config.SaveConfigFile(targetPath, sampleOpts); err != nil {
			return err
		}

		return printer.Success(map[string]any{
			"status":  "created",
			"path":    absPath,
			"message": fmt.Sprintf("配置文件模板已成功初始化于: %s", absPath),
		})
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "查看当前生效的配置参数与会话信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		maskedPassword := ""
		if runtimeOpts.Password != "" {
			maskedPassword = "******"
		}
		safeOpts := map[string]any{
			"configFile": runtimeOpts.ConfigFile,
			"url":        runtimeOpts.URL,
			"account":    runtimeOpts.Account,
			"password":   maskedPassword,
			"accessMode": runtimeOpts.AccessMode,
			"output":     runtimeOpts.Output,
			"timeout":    runtimeOpts.Timeout,
			"insecure":   runtimeOpts.Insecure,
		}
		return printer.Success(safeOpts)
	},
}

func init() {
	configInitCmd.Flags().StringVar(&configInitPath, "path", "", "初始化的配置文件目标路径 (默认: 当前目录 .zentao.yaml)")
	configInitCmd.Flags().BoolVarP(&configInitGlobal, "global", "g", false, "初始化到全局配置路径 ~/.config/zentao/config.yaml")
	configInitCmd.Flags().BoolVarP(&configInitForce, "force", "f", false, "如果配置文件已存在则强制覆盖")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}
