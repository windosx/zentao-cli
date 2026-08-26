package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/windosx/zentao-cli/pkg/zentao"
	"gopkg.in/yaml.v3"
)

// Options holds runtime options gathered from flags, env vars, and config files.
type Options struct {
	ConfigFile string `json:"-" yaml:"-"`
	URL        string `json:"url" yaml:"url"`
	Account    string `json:"account" yaml:"account"`
	Password   string `json:"password" yaml:"password"`
	AccessMode string `json:"accessMode" yaml:"accessMode"`
	Insecure   bool   `json:"insecure" yaml:"insecure"`
	Output     string `json:"output" yaml:"output"`
	Timeout    string `json:"timeout" yaml:"timeout"`
}

// SessionCache holds cached session data to avoid redundant logins.
type SessionCache struct {
	URL       string    `json:"url"`
	Account   string    `json:"account"`
	Cookie    string    `json:"cookie"`
	Rand      string    `json:"rand"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Load merges configurations from config file, environment variables, and flag overrides.
func Load(flagOpts Options) (*Options, error) {
	opts := Options{
		AccessMode: zentao.AccessModeGET,
		Output:     "json",
		Timeout:    "30s",
	}

	// 1. Try loading from file
	if err := loadInitialFile(flagOpts.ConfigFile, &opts); err != nil {
		return nil, err
	}

	// 2. Override with Environment Variables
	applyEnvOverrides(&opts)

	// 3. Override with explicit CLI flags
	applyFlagOverrides(flagOpts, &opts)

	// 4. If URL or Account are still empty, populate from active session.json
	if cache, err := ReadSessionCache("", opts.URL, opts.Account); err == nil && cache != nil {
		if opts.URL == "" && cache.URL != "" {
			opts.URL = cache.URL
		}
		if opts.Account == "" && cache.Account != "" {
			opts.Account = cache.Account
		}
	}

	return &opts, nil
}

func loadInitialFile(configPath string, opts *Options) error {
	if configPath == "" {
		configPath = findDefaultConfigFile()
	}
	if configPath != "" {
		if err := loadFromFile(configPath, opts); err != nil {
			return fmt.Errorf("load config file %q: %w", configPath, err)
		}
		opts.ConfigFile = configPath
	}
	return nil
}

func applyEnvOverrides(opts *Options) {
	if v := os.Getenv("ZENTAO_URL"); v != "" {
		opts.URL = v
	}
	if v := os.Getenv("ZENTAO_ACCOUNT"); v != "" {
		opts.Account = v
	}
	if v := os.Getenv("ZENTAO_PASSWORD"); v != "" {
		opts.Password = v
	}
	if v := os.Getenv("ZENTAO_ACCESS_MODE"); v != "" {
		opts.AccessMode = v
	}
	if v := os.Getenv("ZENTAO_OUTPUT"); v != "" {
		opts.Output = v
	}
	if v := os.Getenv("ZENTAO_TIMEOUT"); v != "" {
		opts.Timeout = v
	}
	if v := os.Getenv("ZENTAO_INSECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			opts.Insecure = b
		}
	}
}

func applyFlagOverrides(flagOpts Options, opts *Options) {
	if flagOpts.URL != "" {
		opts.URL = flagOpts.URL
	}
	if flagOpts.Account != "" {
		opts.Account = flagOpts.Account
	}
	if flagOpts.Password != "" {
		opts.Password = flagOpts.Password
	}
	if flagOpts.AccessMode != "" {
		opts.AccessMode = flagOpts.AccessMode
	}
	if flagOpts.Output != "" {
		opts.Output = flagOpts.Output
	}
	if flagOpts.Timeout != "" {
		opts.Timeout = flagOpts.Timeout
	}
	if flagOpts.Insecure {
		opts.Insecure = true
	}
}

// ToZentaoConfig converts Options into a zentao.Config struct.
func (o *Options) ToZentaoConfig() zentao.Config {
	dur, err := time.ParseDuration(o.Timeout)
	if err != nil || dur <= 0 {
		dur = 30 * time.Second
	}
	return zentao.Config{
		URL:        strings.TrimRight(o.URL, "/"),
		Account:    o.Account,
		Password:   o.Password,
		AccessMode: o.AccessMode,
		Timeout:    dur,
		Insecure:   o.Insecure,
	}
}

// DefaultConfigDir returns ~/.config/zentao.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/zentao"
	}
	return filepath.Join(home, ".config", "zentao")
}

// SaveConfigFile writes the current options to a YAML configuration file.
func SaveConfigFile(path string, opts Options) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(opts)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func findDefaultConfigFile() string {
	cwd, err := os.Getwd()
	if err == nil {
		localYAML := filepath.Join(cwd, ".zentao.yaml")
		if fileExists(localYAML) {
			return localYAML
		}
		localYML := filepath.Join(cwd, ".zentao.yml")
		if fileExists(localYML) {
			return localYML
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		configYAML := filepath.Join(home, ".config", "zentao", "config.yaml")
		if fileExists(configYAML) {
			return configYAML
		}
		configYML := filepath.Join(home, ".config", "zentao", "config.yml")
		if fileExists(configYML) {
			return configYML
		}
		homeYAML := filepath.Join(home, ".zentao.yaml")
		if fileExists(homeYAML) {
			return homeYAML
		}
	}

	return ""
}

func loadFromFile(path string, opts *Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".json") {
		return json.Unmarshal(data, opts)
	}
	return yaml.Unmarshal(data, opts)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
