package skills

import (
	"embed"
)

// FS embeds all skill definitions and references under the skills/ directory.
//
//go:embed */SKILL.md *
var FS embed.FS
