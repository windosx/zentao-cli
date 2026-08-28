package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func TestConfigCommands_InitAndShow(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	customCfgPath := filepath.Join(tempHome, "custom-config.yaml")

	// 1. config init to custom path
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"config", "init", "--path", customCfgPath, "--force", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	var initResp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &initResp); err != nil || initResp["ok"] != true {
		t.Fatalf("unexpected config init response: %s", buf.String())
	}

	// 2. config show with custom path
	buf.Reset()
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"config", "show", "--config", customCfgPath, "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	var showResp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &showResp); err != nil || showResp["ok"] != true {
		t.Fatalf("unexpected config show response: %s", buf.String())
	}
}
