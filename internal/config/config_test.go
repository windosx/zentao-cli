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

func TestStore_ProfileManagement(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// 1. Save Profile
	p1 := Profile{
		Name:       "dev",
		URL:        "http://zentao.dev",
		Account:    "devuser",
		Password:   "devpass",
		Cookie:     "zentaosid=dev123",
		Rand:       "rand1",
		AccessMode: "GET",
	}
	if err := SaveProfile(p1); err != nil {
		t.Fatalf("SaveProfile p1 failed: %v", err)
	}

	p2 := Profile{
		Name:       "prod",
		URL:        "http://zentao.prod",
		Account:    "produser",
		Password:   "prodpass",
		AccessMode: "PATH_INFO",
	}
	if err := SaveProfile(p2); err != nil {
		t.Fatalf("SaveProfile p2 failed: %v", err)
	}

	// 2. Active Profile should be p2 (last saved)
	active, err := GetActiveProfile("")
	if err != nil || active == nil || active.Name != "prod" {
		t.Fatalf("unexpected active profile: %+v, err: %v", active, err)
	}

	// 3. Switch Profile
	if _, err := SwitchProfile("dev"); err != nil {
		t.Fatalf("SwitchProfile to dev failed: %v", err)
	}
	active, err = GetActiveProfile("")
	if err != nil || active == nil || active.Name != "dev" {
		t.Fatalf("unexpected active profile after switch: %+v", active)
	}

	// 4. Switch non-existent
	if _, err := SwitchProfile("notfound"); err == nil {
		t.Errorf("expected error switching to non-existent profile, got nil")
	}

	// 5. UpdateActiveProfileCookie
	UpdateActiveProfileCookie("zentaosid=updated123", "updatedrand")
	active, _ = GetActiveProfile("")
	if active.Cookie != "zentaosid=updated123" || active.Rand != "updatedrand" {
		t.Errorf("cookie update failed: %+v", active)
	}

	// 6. Delete Profile
	if err := DeleteProfile("dev"); err != nil {
		t.Fatalf("DeleteProfile dev failed: %v", err)
	}
	// After deleting active, active should switch to prod
	active, err = GetActiveProfile("")
	if err != nil || active == nil || active.Name != "prod" {
		t.Fatalf("unexpected active profile after delete: %+v", active)
	}

	// 7. Delete non-existent
	if err := DeleteProfile("notfound"); err == nil {
		t.Errorf("expected error deleting non-existent profile")
	}

	// 8. Delete last profile
	if err := DeleteProfile("prod"); err != nil {
		t.Fatalf("DeleteProfile prod failed: %v", err)
	}
	store, err := LoadStore()
	if err != nil || len(store.Profiles) != 0 || store.ActiveProfile != "" {
		t.Errorf("expected empty store after deleting all profiles, got %+v", store)
	}
}

func TestConfig_SaveAndOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "sub", "config.yaml")

	opts := Options{
		ConfigFile: cfgPath,
		URL:        "http://override.test",
		Account:    "admin",
		Insecure:   true,
		Timeout:    "10s",
		AccessMode: "PATH_INFO",
		Output:     "table",
	}

	if err := SaveConfigFile(cfgPath, opts); err != nil {
		t.Fatalf("SaveConfigFile failed: %v", err)
	}

	loaded, err := Load(Options{ConfigFile: cfgPath})
	if err != nil {
		t.Fatalf("Load saved config failed: %v", err)
	}
	if !loaded.Insecure || loaded.Timeout != "10s" || loaded.AccessMode != "PATH_INFO" || loaded.Output != "table" {
		t.Errorf("unexpected loaded options: %+v", loaded)
	}

	// Test flag overrides
	flagOpts := Options{
		ConfigFile: cfgPath,
		Insecure:   false,
		Timeout:    "5s",
		AccessMode: "GET",
		Output:     "json",
	}
	loaded, err = Load(flagOpts)
	if err != nil {
		t.Fatalf("Load with flag overrides failed: %v", err)
	}
	if loaded.Timeout != "5s" || loaded.AccessMode != "GET" || loaded.Output != "json" {
		t.Errorf("flag overrides failed: %+v", loaded)
	}

	// Test env overrides
	t.Setenv("ZENTAO_INSECURE", "true")
	t.Setenv("ZENTAO_TIMEOUT", "20s")
	t.Setenv("ZENTAO_ACCESS_MODE", "PATH_INFO")
	t.Setenv("ZENTAO_OUTPUT", "text")
	loaded, err = Load(Options{ConfigFile: cfgPath})
	if err != nil {
		t.Fatalf("Load with env overrides failed: %v", err)
	}
	if !loaded.Insecure || loaded.Timeout != "20s" || loaded.AccessMode != "PATH_INFO" || loaded.Output != "text" {
		t.Errorf("env overrides failed: %+v", loaded)
	}
}
