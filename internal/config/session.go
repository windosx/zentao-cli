package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultSessionCacheFile returns the path where session token is cached: ~/.config/zentao/session.json.
func DefaultSessionCacheFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "zentao-session.json")
	}
	return filepath.Join(home, ".config", "zentao", "session.json")
}

func fallbackSessionCacheFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "zentao-cli", "session.json")
}

// ReadSessionCache reads cached session from file.
// If url or account are empty, it matches the active cached session.
func ReadSessionCache(cachePath, url, account string) (*SessionCache, error) {
	if cachePath == "" {
		cachePath = DefaultSessionCacheFile()
	}
	data, err := os.ReadFile(cachePath)
	if err != nil && os.IsNotExist(err) {
		if fallback := fallbackSessionCacheFile(); fallback != "" && fallback != cachePath {
			data, err = os.ReadFile(fallback)
		}
	}
	if err != nil {
		return nil, err
	}
	var cache SessionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	if cache.Cookie == "" {
		return nil, os.ErrNotExist
	}

	// Invalidate cache if older than 7 days
	if cache.UpdatedAt.IsZero() || time.Since(cache.UpdatedAt) > 7*24*time.Hour {
		return nil, os.ErrNotExist
	}

	// If url is explicitly provided, verify it matches (ignoring trailing slashes and case)
	if url != "" && cache.URL != "" {
		normA := strings.TrimRight(strings.ToLower(cache.URL), "/")
		normB := strings.TrimRight(strings.ToLower(url), "/")
		if normA != normB {
			return nil, os.ErrNotExist
		}
	}

	// If account is explicitly provided and doesn't match cache
	if account != "" && cache.Account != "" && account != cache.Account {
		return nil, os.ErrNotExist
	}

	return &cache, nil
}

// WriteSessionCache writes session to disk cache.
func WriteSessionCache(cachePath string, cache SessionCache) error {
	if cachePath == "" {
		cachePath = DefaultSessionCacheFile()
	}
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	cache.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0600)
}

// ClearSessionCache removes the cached session.
func ClearSessionCache(cachePath string) error {
	if cachePath == "" {
		cachePath = DefaultSessionCacheFile()
	}
	err := os.Remove(cachePath)
	if fallback := fallbackSessionCacheFile(); fallback != "" && fallback != cachePath {
		_ = os.Remove(fallback)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
