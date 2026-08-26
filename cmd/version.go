package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the current release version of zentao-cli.
	Version = "1.0.0"
	// SDKVersion is the compatible ZenTao PMS SDK release version.
	SDKVersion = "zentaopms_21.7_20250516"
	// GitCommit is set at compile time.
	GitCommit = "HEAD"
	// BuildDate is set at compile time.
	BuildDate = "2026-08-26"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看版本与构建信息",
	Long:  "输出 zentao-cli 版本号、兼容的禅道官方 SDK 版本、Git Commit 及构建环境信息。",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]any{
			"version":    Version,
			"sdkVersion": SDKVersion,
			"gitCommit":  GitCommit,
			"buildDate":  BuildDate,
			"goVersion":  runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		}
		return printer.Success(info)
	},
}
