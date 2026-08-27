package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var schemaCmd = &cobra.Command{
	Use:   "schema [command]",
	Short: "探测命令与参数元数据 Schema（专供 AI Agent 工具发现）",
	Long:  "以结构化 JSON 输出命令定义、必填/选填参数、数据类型、操作风险/副作用与使用示例，支持 AI Agent 进行动态 Tool Calling 发现。",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetCmd := RootCmd
		if len(args) > 0 {
			targetCmd = findSubcommand(RootCmd, args)
			if targetCmd == nil {
				return printer.Success(map[string]any{
					"error": "command not found",
				})
			}
		}

		schema := buildCommandSchema(targetCmd, schemaCompact)
		return printer.Success(schema)
	},
}

var schemaCompact bool

func init() {
	schemaCmd.Flags().BoolVar(&schemaCompact, "compact", false, "返回紧凑模式 Schema (优化 Token 开销)")
}

// CommandSchema represents a tool specification for LLM Agent discovery.
type CommandSchema struct {
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	Description string          `json:"description"`
	Effect      string          `json:"effect"` // read | write | destructive
	Risk        string          `json:"risk"`   // low | medium | high
	UseWhen     string          `json:"use_when,omitempty"`
	AvoidWhen   string          `json:"avoid_when,omitempty"`
	Parameters  []ParamSchema   `json:"parameters,omitempty"`
	Subcommands []CommandSchema `json:"subcommands,omitempty"`
	Examples    []string        `json:"examples,omitempty"`
}

// ParamSchema describes a CLI parameter flag.
type ParamSchema struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

func buildCommandSchema(cmd *cobra.Command, compact bool) CommandSchema {
	effect := "read"
	risk := "low"
	name := cmd.Name()

	if strings.Contains(name, "create") || strings.Contains(name, "add") || strings.Contains(name, "finish") || strings.Contains(name, "resolve") || strings.Contains(name, "login") || strings.Contains(name, "start") || strings.Contains(name, "close") {
		effect = "write"
		risk = "medium"
	}
	if strings.Contains(name, "params") {
		effect = "read"
		risk = "low"
	}
	if strings.Contains(name, "delete") || strings.Contains(name, "remove") || strings.Contains(name, "logout") {
		effect = "destructive"
		risk = "high"
	}

	cs := CommandSchema{
		Name:        cmd.Name(),
		Path:        cmd.CommandPath(),
		Description: cmd.Short,
		Effect:      effect,
		Risk:        risk,
	}

	// Flags
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Filter out standard help flags
		if f.Name == "help" {
			return
		}
		ps := ParamSchema{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Required:    strings.Contains(f.Usage, "(必填)") || strings.Contains(f.Usage, "(required)"),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		cs.Parameters = append(cs.Parameters, ps)
	})

	// Subcommands
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		if compact {
			cs.Subcommands = append(cs.Subcommands, CommandSchema{
				Name:        sub.Name(),
				Path:        sub.CommandPath(),
				Description: sub.Short,
			})
		} else {
			cs.Subcommands = append(cs.Subcommands, buildCommandSchema(sub, compact))
		}
	}

	return cs
}

func findSubcommand(root *cobra.Command, path []string) *cobra.Command {
	curr := root
	for _, name := range path {
		found := false
		for _, sub := range curr.Commands() {
			if sub.Name() == name {
				curr = sub
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return curr
}
