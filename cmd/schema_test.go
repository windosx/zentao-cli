package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func TestSchemaCommands_FullAndCompact(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// 1. schema for whole root in compact format
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"schema", "--compact", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("schema --compact failed: %v", err)
	}

	var respCompact map[string]any
	if err := json.Unmarshal(buf.Bytes(), &respCompact); err != nil || respCompact["ok"] != true {
		t.Fatalf("unexpected schema response: %s", buf.String())
	}

	// 2. schema for specific module "task"
	buf.Reset()
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"schema", "task", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("schema task failed: %v", err)
	}

	var respTask map[string]any
	if err := json.Unmarshal(buf.Bytes(), &respTask); err != nil || respTask["ok"] != true {
		t.Fatalf("unexpected schema task response: %s", buf.String())
	}
}
