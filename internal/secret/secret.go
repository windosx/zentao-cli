// Package secret provides secure credential storage backed by the OS keychain
// (macOS Keychain, Windows Credential Manager, Linux Secret Service) with
// automatic graceful fallback to a permission-restricted local file in headless,
// CI, test, or permission-denied environments.
package secret

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/windosx/zentao-cli/internal/config"
)

const service = "zentao-cli"

// isTestEnv checks whether the current process is running under `go test`.
func isTestEnv() bool {
	if os.Getenv("ZENTAO_NO_KEYRING") == "1" {
		return true
	}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

func fileRing() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName:     service,
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         filepath.Join(config.DefaultConfigDir(), "keyring-file"),
		FilePasswordFunc: func(string) (string, error) {
			return "", nil
		},
	})
}

func getRing() (keyring.Keyring, error) {
	if isTestEnv() {
		return fileRing()
	}

	return keyring.Open(keyring.Config{
		ServiceName: service,
		FileDir:     filepath.Join(config.DefaultConfigDir(), "keyring-file"),
		FilePasswordFunc: func(string) (string, error) {
			return "", nil
		},
	})
}

// Set stores a password in the system keyring, falling back to local file storage on failure.
func Set(url, account, password string) error {
	key := fmt.Sprintf("%s@%s", account, url)
	ring, err := getRing()
	if err == nil {
		if err = ring.Set(keyring.Item{Key: key, Data: []byte(password)}); err == nil {
			return nil
		}
	}

	// Fallback to file backend if native keychain fails/denied
	fallback, fErr := fileRing()
	if fErr != nil {
		return fErr
	}
	return fallback.Set(keyring.Item{Key: key, Data: []byte(password)})
}

// Get retrieves a password from the system keyring, falling back to local file storage on failure.
func Get(url, account string) (string, error) {
	key := fmt.Sprintf("%s@%s", account, url)
	ring, err := getRing()
	if err == nil {
		if item, err := ring.Get(key); err == nil && len(item.Data) > 0 {
			return string(item.Data), nil
		}
	}

	// Fallback to file backend
	fallback, fErr := fileRing()
	if fErr != nil {
		return "", fErr
	}
	item, err := fallback.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

// Delete removes a stored password from the system keyring and file storage.
func Delete(url, account string) error {
	key := fmt.Sprintf("%s@%s", account, url)
	ring, err := getRing()
	if err == nil {
		_ = ring.Remove(key)
	}
	fallback, fErr := fileRing()
	if fErr == nil {
		_ = fallback.Remove(key)
	}
	return nil
}
