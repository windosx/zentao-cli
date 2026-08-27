package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/windosx/zentao-cli/skills"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "管理与安装 AI Agent Skill (ZCode, Claude, Cursor, Agents 等)",
	Long:  "为各大 AI Agent 平台自动安装或分发标准化的 SKILL.md 技能定义，实现零配置开箱即用。",
}

var (
	skillSetupTarget string
)

var skillSetupCmd = &cobra.Command{
	Use:     "setup",
	Aliases: []string{"update", "sync"},
	Short:   "安装或更新 zentao Agent Skill 到本地 AI Agent 运行环境",
	Long:    "自动将内置最新的 SKILL.md 分发并写入 ~/.zcode/skills/zentao/、~/.agents/skills/zentao/ 与 ~/.claude/skills/zentao/ 目录。",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		installedPaths := []string{}
		targets := []string{}

		switch skillSetupTarget {
		case "zcode":
			targets = append(targets, filepath.Join(home, ".zcode", "skills", "zentao", "SKILL.md"))
		case "agents":
			targets = append(targets, filepath.Join(home, ".agents", "skills", "zentao", "SKILL.md"))
		case "claude":
			targets = append(targets, filepath.Join(home, ".claude", "skills", "zentao", "SKILL.md"))
		default: // "all"
			targets = append(targets,
				filepath.Join(home, ".zcode", "skills", "zentao", "SKILL.md"),
				filepath.Join(home, ".agents", "skills", "zentao", "SKILL.md"),
				filepath.Join(home, ".claude", "skills", "zentao", "SKILL.md"),
			)
		}

		for _, p := range targets {
			dir := filepath.Dir(p)
			if err := os.MkdirAll(dir, 0700); err != nil {
				continue
			}

			// Use atomic copy from embedded FS via io.Copy
			if err := copySkillFile("zentao/SKILL.md", p); err == nil {
				installedPaths = append(installedPaths, p)
			}
		}

		return printer.Success(map[string]any{
			"status":         "installed",
			"installedPaths": installedPaths,
			"message":        fmt.Sprintf("已成功安装/更新 zentao Agent Skill 到 %d 个目标环境", len(installedPaths)),
		})
	},
}

// AutoSyncInstalledSkills silently checks existing installed SKILL.md files in user's home directory.
// If any installed file differs from embedded FS, it atomically overwrites it with the latest version.
func AutoSyncInstalledSkills() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	embeddedData, err := skills.FS.ReadFile("zentao/SKILL.md")
	if err != nil || len(embeddedData) == 0 {
		return
	}

	candidates := []string{
		filepath.Join(home, ".zcode", "skills", "zentao", "SKILL.md"),
		filepath.Join(home, ".agents", "skills", "zentao", "SKILL.md"),
		filepath.Join(home, ".claude", "skills", "zentao", "SKILL.md"),
	}

	for _, p := range candidates {
		// Only sync if the file already exists (i.e. user previously installed or uses this path)
		existing, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		if !bytes.Equal(existing, embeddedData) {
			_ = copySkillFile("zentao/SKILL.md", p)
		}
	}
}

func copySkillFile(embeddedRelPath, destPath string) error {
	src, err := skills.FS.Open(embeddedRelPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	tmpPath := fmt.Sprintf("%s.tmp", destPath)
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		_ = dst.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	if err := dst.Sync(); err != nil {
		return err
	}
	_ = dst.Close()

	return os.Rename(tmpPath, destPath)
}

func init() {
	skillSetupCmd.Flags().StringVar(&skillSetupTarget, "target", "all", "安装/更新目标环境: all (全部), zcode, agents, claude")

	skillCmd.AddCommand(skillSetupCmd)
}
