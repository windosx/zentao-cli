package secret

import (
	"testing"
)

func TestSecret_CRUD(t *testing.T) {
	t.Setenv("ZENTAO_NO_KEYRING", "1")
	t.Setenv("HOME", t.TempDir())

	url := "http://example.com"
	account := "admin"
	password := "Secret123"

	// 1. Set
	if err := Set(url, account, password); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 2. Get
	got, err := Get(url, account)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != password {
		t.Errorf("Get() = %q, want %q", got, password)
	}

	// 3. Delete
	if err := Delete(url, account); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 4. Verify Delete
	_, err = Get(url, account)
	if err == nil {
		t.Error("expected error getting deleted secret, got nil")
	}
}
