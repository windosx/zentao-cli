package cmd

import (
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Version is the current release version of zentao-cli.
	// 由 goreleaser 通过 -X 注入；go install 构建时回退到 build info 中的模块版本。
	Version = "dev"
	// SDKVersion is the compatible ZenTao PMS SDK release version.
	SDKVersion = "zentaopms_21.7_20250516"
	// ZenTaoCompat is the target ZenTao PMS major/minor version compatibility.
	ZenTaoCompat = "21.7"
	// GitCommit is set at compile time.
	GitCommit = "HEAD"
	// BuildDate is set at compile time.
	BuildDate = "unknown"
)

// resolveVersion 返回实际发布版本：
//  1. goreleaser 注入的 Version；
//  2. go install pkg@vX.Y.Z 时从 build info 读取模块版本（带 v 前缀）。
func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/windosx/zentao-cli" && dep.Version != "" {
				return strings.TrimPrefix(dep.Version, "v")
			}
		}
		// 主模块本身即 zentao-cli 时，Main.Version 即版本
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return Version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看版本与构建信息",
	Long:  "输出 zentao-cli 版本号、兼容的禅道官方 SDK 版本、Git Commit 及构建环境信息。",
	RunE: func(cmd *cobra.Command, args []string) error {
		ver := resolveVersion()
		fullVer := ver
		if !strings.Contains(ver, "+") {
			fullVer = ver + "+" + ZenTaoCompat
		}

		info := map[string]any{
			"version":      ver,
			"fullVersion":  fullVer,
			"zentaoCompat": "v" + ZenTaoCompat + "+",
			"sdkVersion":   SDKVersion,
			"gitCommit":    GitCommit,
			"buildDate":    BuildDate,
			"goVersion":    runtime.Version(),
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
		}
		return printer.Success(info)
	},
}

func init() {
	RootCmd.Version = resolveVersion()
}
