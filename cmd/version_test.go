package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func TestVersionCommands_ExecutionAndFlags(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// 1. zentao version -o json
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"version", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
		t.Fatalf("unexpected version response: %s", buf.String())
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map in version response, got: %v", resp)
	}
	if data["sdkVersion"] != SDKVersion {
		t.Errorf("expected sdkVersion=%s, got %v", SDKVersion, data["sdkVersion"])
	}
	if data["zentaoCompat"] != "v"+ZenTaoCompat+"+" {
		t.Errorf("expected zentaoCompat=v%s+, got %v", ZenTaoCompat, data["zentaoCompat"])
	}

	// 2. zentao --version flag
	buf.Reset()
	flagOpts = config.Options{}
	if hFlag := RootCmd.Flags().Lookup("help"); hFlag != nil {
		hFlag.Changed = false
		_ = hFlag.Value.Set("false")
	}
	if vFlag := RootCmd.Flags().Lookup("version"); vFlag != nil {
		vFlag.Changed = false
		_ = vFlag.Value.Set("false")
	}
	RootCmd.SetArgs([]string{"--version"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("zentao --version failed: %v", err)
	}
	if !strings.Contains(buf.String(), "zentao version") && !strings.Contains(buf.String(), resolveVersion()) {
		t.Errorf("expected version output from --version flag, got: %s", buf.String())
	}

	// 3. resolveVersion fallback logic test
	ver := resolveVersion()
	if ver == "" {
		t.Errorf("resolveVersion returned empty string")
	}
}
