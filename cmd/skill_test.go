package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func TestSkillCommands_SetupAndSync(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// 1. Test skill setup with target=agents
	flagOpts = config.Options{}
	skillSetupTarget = "agents"
	RootCmd.SetArgs([]string{"skill", "setup", "--target", "agents", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("skill setup failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
		t.Fatalf("unexpected skill setup response: %s", buf.String())
	}

	targetSkillPath := filepath.Join(tempHome, ".agents", "skills", "zentao", "SKILL.md")
	if _, err := os.Stat(targetSkillPath); os.IsNotExist(err) {
		t.Fatalf("expected installed skill file at %s", targetSkillPath)
	}

	// 2. Test skill sync alias
	buf.Reset()
	flagOpts = config.Options{}
	skillSetupTarget = "all"
	RootCmd.SetArgs([]string{"skill", "sync", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("skill sync failed: %v", err)
	}

	// 3. Test AutoSyncInstalledSkills
	AutoSyncInstalledSkills()
}
