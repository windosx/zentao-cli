package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/internal/output"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("ZENTAO_NO_KEYRING", "1")
	os.Exit(m.Run())
}

func TestRootCmd_Help(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"--help"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing root help: %v", err)
	}
	if hFlag := RootCmd.Flags().Lookup("help"); hFlag != nil {
		hFlag.Changed = false
		_ = hFlag.Value.Set("false")
	}

	out := buf.String()
	subcommands := []string{
		"auth", "my", "todo", "task", "bug", "story",
		"project", "product", "user", "dept", "trash",
		"config", "schema", "skill", "version",
	}
	for _, sub := range subcommands {
		if !bytes.Contains([]byte(out), []byte(sub)) {
			t.Errorf("expected subcommand %q in root help output", sub)
		}
	}
}

func TestTransparentAutoRelogin_WhenSessionExpires(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var loginCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"fresh-session-456","rand":"rand-789"}`,
			})
		case m == "user" && f == "login":
			atomic.AddInt32(&loginCount, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "testuser"},
			})
		case m == "my" && f == "task":
			cookie := r.Header.Get("Cookie")
			if cookie == "zentaosid=stale-session-123" {
				// Simulate ZenTao PHP server session timeout response
				_ = json.NewEncoder(w).Encode(map[string]any{
					"result":  false,
					"message": "登录已超时，请重新登入!",
					"load":    "login",
				})
				return
			}
			if cookie == "zentaosid=fresh-session-456" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"data":   `[{"id":"101","name":"Recovered Task","status":"doing"}]`,
				})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 1. First auth login
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"auth", "login", "-u", server.URL, "-a", "testuser", "-p", "Test@123456", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("initial auth login failed: %v", err)
	}

	// 2. Manually corrupt the cached cookie to simulate server-side session expiration
	// Note: Password is intentionally empty in Profile to verify Keyring password resolution!
	_ = config.SaveProfile(config.Profile{
		URL:        server.URL,
		Account:    "testuser",
		Password:   "",
		Cookie:     "zentaosid=stale-session-123",
		AccessMode: "GET",
	})

	// 3. Execute 'zentao my task' - it should transparently detect session timeout, re-login, and succeed
	buf.Reset()
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"my", "task", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("my task with expired session should have auto-recovered, but failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
		t.Fatalf("expected recovered successful response, got: %s", buf.String())
	}

	// Verify that login was called again for auto-recovery
	if atomic.LoadInt32(&loginCount) < 2 {
		t.Errorf("expected login to be re-invoked automatically, loginCount=%d", atomic.LoadInt32(&loginCount))
	}
}

func TestClassifyError_And_FrameworkValidation(t *testing.T) {
	// 1. Direct classification mapping
	if code, cat := classifyError(nil); code != output.ExitCodeSuccess || cat != "none" {
		t.Errorf("unexpected success classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(&zentao.Error{Kind: zentao.KindAuth, Message: "session timeout"}); code != output.ExitCodeAuth || cat != "auth" {
		t.Errorf("unexpected auth classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("%w: invalid param", zentao.ErrValidation)); code != output.ExitCodeValidation || cat != "validation" {
		t.Errorf("unexpected validation classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("unknown flag: --foo")); code != output.ExitCodeValidation || cat != "validation" {
		t.Errorf("unexpected cobra usage classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("api internal error")); code != output.ExitCodeAPI || cat != "api" {
		t.Errorf("unexpected api classification: code=%d, cat=%s", code, cat)
	}

	// 2. Unauthenticated execution check
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"my", "task", "-o", "json"})
	if err := RootCmd.Execute(); err == nil {
		t.Fatalf("expected unauthenticated error when running my task without login, got nil")
	}

	// 3. Direct ensureClientLoggedIn check
	client = nil
	if err := ensureClientLoggedIn(t.Context()); err == nil {
		t.Errorf("expected error when client is nil")
	}
}

func TestBindProfileToClient(t *testing.T) {
	origClient := client
	defer func() { client = origClient }()

	client = zentao.New(zentao.Config{})
	p := &config.Profile{
		URL:      "http://example.com/",
		Account:  "admin",
		Password: "password123",
		Cookie:   "zentaosid=123",
		Rand:     "rand456",
	}

	bindProfileToClient(p)

	if client.BaseURL != "http://example.com" {
		t.Errorf("expected BaseURL without trailing slash, got: %s", client.BaseURL)
	}
	if client.Account != "admin" {
		t.Errorf("expected Account admin, got: %s", client.Account)
	}
	if client.Password != "password123" {
		t.Errorf("expected Password password123, got: %s", client.Password)
	}
	if client.Cookie != "zentaosid=123" {
		t.Errorf("expected Cookie zentaosid=123, got: %s", client.Cookie)
	}
	if client.GetRand() != "rand456" {
		t.Errorf("expected Rand rand456, got: %s", client.GetRand())
	}
}
