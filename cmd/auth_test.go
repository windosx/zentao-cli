package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func TestMaskCookie(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"zentaosid=abcdef123456", "zentaosid=****3456"},
		{"zentaosid=abcd", "zentaosid=****abcd"},
		{"novalue", "****alue"},
	}
	for _, tt := range tests {
		if got := maskCookie(tt.in); got != tt.want {
			t.Errorf("maskCookie(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAuthCommands_FullSuite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"auth-sess-1","rand":"auth-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{}`})
		}
	}))
	defer server.Close()

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// resetAuthFlags resets auth login package-level flag variables to their
	// declared defaults so prior tests that called auth login cannot bleed
	// a non-empty password (or other field) into this test's validation cases.
	resetAuthFlags := func() {
		loginURL = ""
		loginAccount = ""
		loginPassword = ""
		loginAccessMode = authLoginCmd.Flags().Lookup("access-mode").DefValue
		profileName = ""
		showSecrets = false
	}

	// run executes a command and returns its error. Each call resets the shared
	// buffer, flagOpts, and auth login flags so sub-tests are isolated from one
	// another regardless of execution order (including -shuffle=on).
	run := func(t *testing.T, args []string) error {
		t.Helper()
		buf.Reset()
		flagOpts = config.Options{}
		resetAuthFlags()
		RootCmd.SetArgs(args)
		return RootCmd.Execute()
	}

	t.Run("validation: login missing password", func(t *testing.T) {
		if err := run(t, []string{"auth", "login", "-u", server.URL, "-a", "admin", "-o", "json"}); err == nil {
			t.Fatal("expected error when login without password")
		}
	})

	t.Run("login with custom name", func(t *testing.T) {
		if err := run(t, []string{"auth", "login", "-u", server.URL, "-a", "admin", "-p", "123456", "--name", "prod", "-o", "json"}); err != nil {
			t.Fatalf("auth login failed: %v", err)
		}
	})

	t.Run("auth list", func(t *testing.T) {
		if err := run(t, []string{"auth", "list", "-o", "json"}); err != nil {
			t.Fatalf("auth list failed: %v", err)
		}
	})

	t.Run("auth switch", func(t *testing.T) {
		if err := run(t, []string{"auth", "switch", "--name", "prod", "-o", "json"}); err != nil {
			t.Fatalf("auth switch failed: %v", err)
		}
	})

	t.Run("auth status", func(t *testing.T) {
		if err := run(t, []string{"auth", "status", "--show-secrets", "-o", "json"}); err != nil {
			t.Fatalf("auth status failed: %v", err)
		}
	})

	t.Run("auth logout", func(t *testing.T) {
		if err := run(t, []string{"auth", "logout", "-o", "json"}); err != nil {
			t.Fatalf("auth logout failed: %v", err)
		}
	})

	// Restore a valid session after logout so tests in other files that run
	// after this one (sorted alphabetically: bug, config, dept …) don't start
	// with a nil client. Each of those files performs its own login, but this
	// Cleanup makes the dependency explicit and survivable under -shuffle=on.
	t.Cleanup(func() {
		flagOpts = config.Options{}
		RootCmd.SetArgs([]string{"auth", "login", "-u", server.URL, "-a", "admin", "-p", "123456", "-o", "json"})
		_ = RootCmd.Execute()
	})
}
