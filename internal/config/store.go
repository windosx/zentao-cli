package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Profile represents an authenticated ZenTao account/server instance.
type Profile struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Account    string    `json:"account"`
	Password   string    `json:"password,omitempty"`
	Cookie     string    `json:"cookie"`
	Rand       string    `json:"rand"`
	AccessMode string    `json:"accessMode"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Store holds multiple named profiles and points to the currently active profile.
type Store struct {
	Profiles      map[string]*Profile `json:"profiles"`
	ActiveProfile string              `json:"active_profile"`
}

var storeMu sync.Mutex

// ProfilesFilePath returns ~/.config/zentao/profiles.json.
func ProfilesFilePath() string {
	return filepath.Join(DefaultConfigDir(), "profiles.json")
}

// LoadStore reads profiles from disk. If not found, initializes an empty Store.
func LoadStore() (*Store, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	path := ProfilesFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{
				Profiles: make(map[string]*Profile),
			}, nil
		}
		return nil, err
	}

	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return &Store{Profiles: make(map[string]*Profile)}, nil
	}
	if s.Profiles == nil {
		s.Profiles = make(map[string]*Profile)
	}
	return &s, nil
}

// SaveProfile atomically writes or updates a profile in the store and sets it as active.
func SaveProfile(profile Profile) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	s, err := loadStoreUnsafe()
	if err != nil {
		s = &Store{Profiles: make(map[string]*Profile)}
	}

	name := profile.Name
	if name == "" {
		// Generate default profile name: account@host
		normURL := strings.TrimRight(strings.ToLower(profile.URL), "/")
		normURL = strings.TrimPrefix(normURL, "http://")
		normURL = strings.TrimPrefix(normURL, "https://")
		name = fmt.Sprintf("%s@%s", profile.Account, normURL)
		profile.Name = name
	}

	// Preserve password if not provided during update
	if profile.Password == "" && s.Profiles[name] != nil {
		profile.Password = s.Profiles[name].Password
	}

	profile.UpdatedAt = time.Now()
	s.Profiles[name] = &profile
	s.ActiveProfile = name

	// Atomic save to profiles.json
	if err := saveStoreUnsafe(s); err != nil {
		return err
	}

	// Also sync to session.json
	_ = writeSessionCacheUnsafe(DefaultSessionCacheFile(), SessionCache{
		URL:       profile.URL,
		Account:   profile.Account,
		Cookie:    profile.Cookie,
		Rand:      profile.Rand,
		UpdatedAt: profile.UpdatedAt,
	})

	return nil
}

// UpdateActiveProfileCookie updates only the cookie and rand for active profile after automatic refresh.
func UpdateActiveProfileCookie(cookie, rand string) {
	storeMu.Lock()
	defer storeMu.Unlock()

	s, err := loadStoreUnsafe()
	if err != nil || s.ActiveProfile == "" {
		return
	}
	if p, exists := s.Profiles[s.ActiveProfile]; exists {
		p.Cookie = cookie
		p.Rand = rand
		p.UpdatedAt = time.Now()
		_ = saveStoreUnsafe(s)

		_ = writeSessionCacheUnsafe(DefaultSessionCacheFile(), SessionCache{
			URL:       p.URL,
			Account:   p.Account,
			Cookie:    cookie,
			Rand:      rand,
			UpdatedAt: p.UpdatedAt,
		})
	}
}

// GetActiveProfile returns the current active profile, or nil if none is set.
func GetActiveProfile(specificProfile string) (*Profile, error) {
	s, err := LoadStore()
	if err != nil {
		return nil, err
	}

	if specificProfile != "" {
		p, exists := s.Profiles[specificProfile]
		if !exists {
			return nil, fmt.Errorf("profile %q not found", specificProfile)
		}
		return p, nil
	}

	if s.ActiveProfile != "" {
		if p, exists := s.Profiles[s.ActiveProfile]; exists && (p.Cookie != "" || p.Password != "") {
			return p, nil
		}
	}

	return fallbackSessionProfile()
}

func fallbackSessionProfile() (*Profile, error) {
	cache, err := ReadSessionCache("", "", "")
	if err == nil && cache != nil && cache.Cookie != "" {
		return &Profile{
			Name:       "default",
			URL:        cache.URL,
			Account:    cache.Account,
			Cookie:     cache.Cookie,
			Rand:       cache.Rand,
			AccessMode: "GET",
			UpdatedAt:  cache.UpdatedAt,
		}, nil
	}
	return nil, os.ErrNotExist
}

// SwitchProfile changes the active profile.
func SwitchProfile(name string) (*Profile, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	s, err := loadStoreUnsafe()
	if err != nil {
		return nil, err
	}
	p, exists := s.Profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	s.ActiveProfile = name
	if err := saveStoreUnsafe(s); err != nil {
		return nil, err
	}

	_ = writeSessionCacheUnsafe(DefaultSessionCacheFile(), SessionCache{
		URL:       p.URL,
		Account:   p.Account,
		Cookie:    p.Cookie,
		Rand:      p.Rand,
		UpdatedAt: p.UpdatedAt,
	})

	return p, nil
}

// DeleteProfile removes a profile from the store.
func DeleteProfile(name string) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	s, err := loadStoreUnsafe()
	if err != nil {
		return err
	}
	delete(s.Profiles, name)
	if s.ActiveProfile == name {
		s.ActiveProfile = ""
		for k := range s.Profiles {
			s.ActiveProfile = k
			break
		}
	}

	return saveStoreUnsafe(s)
}

func loadStoreUnsafe() (*Store, error) {
	path := ProfilesFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return &Store{Profiles: make(map[string]*Profile)}, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return &Store{Profiles: make(map[string]*Profile)}, err
	}
	if s.Profiles == nil {
		s.Profiles = make(map[string]*Profile)
	}
	return &s, nil
}

func saveStoreUnsafe(s *Store) error {
	path := ProfilesFilePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write via temp file
	tmpPath := fmt.Sprintf("%s.tmp-%d", path, time.Now().UnixNano())
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func writeSessionCacheUnsafe(cachePath string, cache SessionCache) error {
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cache.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := fmt.Sprintf("%s.tmp-%d", cachePath, time.Now().UnixNano())
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, cachePath)
}
