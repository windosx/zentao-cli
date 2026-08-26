package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Precedence(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "zentao.yaml")

	yamlContent := `
url: "http://from-file.com"
account: "file-user"
password: "file-password"
accessMode: "PATH_INFO"
output: "yaml"
timeout: "45s"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	// 1. Loaded from file
	opts, err := Load(Options{ConfigFile: cfgPath})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if opts.URL != "http://from-file.com" || opts.Account != "file-user" || opts.AccessMode != "PATH_INFO" {
		t.Errorf("unexpected loaded options: %+v", opts)
	}

	// 2. Env overrides file
	t.Setenv("ZENTAO_URL", "http://from-env.com")
	t.Setenv("ZENTAO_ACCOUNT", "env-user")
	opts, err = Load(Options{ConfigFile: cfgPath})
	if err != nil {
		t.Fatalf("Load with env failed: %v", err)
	}
	if opts.URL != "http://from-env.com" || opts.Account != "env-user" {
		t.Errorf("env override failed: %+v", opts)
	}

	// 3. Flags override env
	opts, err = Load(Options{
		ConfigFile: cfgPath,
		URL:        "http://from-flags.com",
		Account:    "flag-user",
	})
	if err != nil {
		t.Fatalf("Load with flags failed: %v", err)
	}
	if opts.URL != "http://from-flags.com" || opts.Account != "flag-user" {
		t.Errorf("flag override failed: %+v", opts)
	}

	// 4. Test ToZentaoConfig
	zcfg := opts.ToZentaoConfig()
	if zcfg.URL != "http://from-flags.com" || zcfg.Account != "flag-user" || zcfg.Timeout != 45*time.Second {
		t.Errorf("unexpected zentao config: %+v", zcfg)
	}
}

func TestSessionCache_ReadWriteClear(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "session.json")

	cache := SessionCache{
		URL:     "http://zentao.test",
		Account: "admin",
		Cookie:  "zentaosid=abc123456",
		Rand:    "987654",
	}

	if err := WriteSessionCache(cachePath, cache); err != nil {
		t.Fatalf("WriteSessionCache failed: %v", err)
	}

	read, err := ReadSessionCache(cachePath, "http://zentao.test", "admin")
	if err != nil {
		t.Fatalf("ReadSessionCache failed: %v", err)
	}
	if read.Cookie != "zentaosid=abc123456" || read.Rand != "987654" {
		t.Errorf("unexpected read session cache: %+v", read)
	}

	// Test mismatched account
	if _, err := ReadSessionCache(cachePath, "http://zentao.test", "other"); err == nil {
		t.Errorf("expected error for mismatched account, got nil")
	}

	// Test clear
	if err := ClearSessionCache(cachePath); err != nil {
		t.Fatalf("ClearSessionCache failed: %v", err)
	}
	if _, err := ReadSessionCache(cachePath, "http://zentao.test", "admin"); err == nil {
		t.Errorf("expected error after clearing session, got nil")
	}
}
