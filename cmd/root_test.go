package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/internal/output"
)

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

	out := buf.String()
	subcommands := []string{"auth", "my", "todo", "dept", "user", "product", "project", "task", "bug", "config", "schema", "skill", "version"}
	for _, sub := range subcommands {
		if !bytes.Contains([]byte(out), []byte(sub)) {
			t.Errorf("expected subcommand %q in help output", sub)
		}
	}
}

func TestVersionCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"version", "-o", "json"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	out := buf.String()
	expectedVersion := resolveVersion()
	if !bytes.Contains([]byte(out), []byte(`"version": "`+expectedVersion+`"`)) {
		t.Errorf("unexpected version output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte(`"sdkVersion": "zentaopms_21.7_20250516"`)) {
		t.Errorf("expected sdkVersion in version output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte(`"zentaoCompat": "v21.7+"`)) {
		t.Errorf("expected zentaoCompat in version output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte(`"fullVersion": "`)) {
		t.Errorf("expected fullVersion in version output: %s", out)
	}
}

func TestSchemaCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"schema", "my", "--compact", "-o", "json"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
		t.Fatalf("unexpected schema response: %s", buf.String())
	}
}

func TestSkillSetupCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"skill", "setup", "--target", "agents", "-o", "json"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("skill setup command failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
		t.Fatalf("unexpected skill setup response: %s", buf.String())
	}
}

func TestConfigInitAndShow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	tmpFile := t.TempDir() + "/test-config.yaml"

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"config", "init", "--path", tmpFile, "--force", "-o", "json"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	buf.Reset()
	flagOpts = config.Options{}
	RootCmd.SetArgs([]string{"config", "show", "--config", tmpFile, "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal config show output: %v", err)
	}
	if resp["ok"] != true {
		t.Errorf("expected ok=true in config show response")
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
	_ = config.SaveProfile(config.Profile{
		URL:        server.URL,
		Account:    "testuser",
		Password:   "Test@123456",
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

func TestAuthLogin_And_SubcommandsFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"test-sess-123","rand":"rand-456"}`,
			})
		case m == "user" && f == "login":
			_ = r.ParseForm()
			if r.FormValue("account") == "testuser" && r.FormValue("password") == "Test@123456" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"user":   map[string]any{"id": "1", "account": "testuser"},
				})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "failed", "reason": "用户名或密码错误"})
			}
		case m == "my" && f == "task":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"101","name":"My Task 1","status":"doing"}]`})
		case m == "my" && f == "bug":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"201","title":"My Bug 1","status":"active"}]`})
		case m == "my" && f == "todo":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"301","name":"My Todo 1","status":"wait"}]`})
		case m == "my" && f == "story":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"401","title":"My Story 1","status":"active"}]`})
		case m == "my" && f == "project":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"501","name":"My Project 1","status":"doing"}]`})
		case m == "my" && f == "dynamic":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"601","actor":"testuser","action":"created","objectType":"task","objectName":"Task 1"}]`})
		case m == "todo" && f == "create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "todo created"})
		case m == "todo" && f == "start":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "todo started"})
		case m == "todo" && f == "finish":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "todo finished"})
		case m == "todo" && f == "close":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "todo closed"})
		case m == "todo" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "todo deleted"})
		case m == "dept" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"Dev"}]`})
		case m == "dept" && f == "manageChild":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "dept added"})
		case m == "company" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","account":"dev1"}]`})
		case m == "user" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"depts":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "user created"})
			}
		case m == "product" && f == "all" || m == "product" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"App"}]`})
		case m == "product" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"products":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "product created"})
			}
		case m == "project" && f == "browse" || m == "project" && f == "all" || m == "execution" && f == "all":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"Sprint 1"}]`})
		case m == "project" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"allProducts":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "project created"})
			}
		case m == "project" && f == "task" || m == "execution" && f == "task":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"Task 1"}]`})
		case m == "task" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"projects":{"1":"Sprint 1"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "task created"})
			}
		case m == "task" && f == "finish":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"task":{"id":"1"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "task finished"})
			}
		case m == "task" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "task deleted"})
		case m == "bug" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","title":"Bug 1"}]`})
		case m == "bug" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"types":{"codeerror":"Code Error"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "bug created"})
			}
		case m == "bug" && f == "resolve":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"bug":{"id":"1"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "bug resolved"})
			}
		case m == "bug" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "bug deleted"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 1. Perform auth login once
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"auth", "login", "-u", server.URL, "-a", "testuser", "-p", "Test@123456", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("auth login failed: %v", err)
	}

	// 2. Execute all subsequent commands WITHOUT passing -a or -p
	testCases := []struct {
		name string
		args []string
	}{
		{"auth status", []string{"auth", "status"}},
		{"auth list", []string{"auth", "list"}},
		{"my task", []string{"my", "task"}},
		{"my bug", []string{"my", "bug"}},
		{"my todo", []string{"my", "todo"}},
		{"my story", []string{"my", "story"}},
		{"my project", []string{"my", "project"}},
		{"my dynamic", []string{"my", "dynamic"}},
		{"todo list", []string{"todo", "list"}},
		{"todo create", []string{"todo", "create", "--name", "Review PR"}},
		{"todo start", []string{"todo", "start", "--id", "301"}},
		{"todo finish", []string{"todo", "finish", "--id", "301"}},
		{"todo close", []string{"todo", "close", "--id", "301"}},
		{"todo delete", []string{"todo", "delete", "--id", "301"}},
		{"dept list", []string{"dept", "list"}},
		{"dept add", []string{"dept", "add", "--parent", "1", "--name", "Frontend"}},
		{"user list", []string{"user", "list"}},
		{"user params", []string{"user", "params", "--dept", "1"}},
		{"user add", []string{"user", "add", "--username", "tom", "--user-password", "pwd123", "--realname", "Tom"}},
		{"product list", []string{"product", "list"}},
		{"product params", []string{"product", "params", "--program", "1"}},
		{"product add", []string{"product", "add", "--name", "Web", "--code", "web"}},
		{"project list", []string{"project", "list"}},
		{"project params", []string{"project", "params", "--program", "1"}},
		{"project add", []string{"project", "add", "--name", "Sprint 2", "--code", "s2"}},
		{"task list", []string{"task", "list", "--project", "1"}},
		{"task params", []string{"task", "params", "--project", "1"}},
		{"task create", []string{"task", "create", "--project", "1", "--name", "Implement Feature"}},
		{"task finish-params", []string{"task", "finish-params", "--id", "1"}},
		{"task finish", []string{"task", "finish", "--id", "1", "--real", "2.0"}},
		{"task delete", []string{"task", "delete", "--project", "1", "--id", "1"}},
		{"bug list", []string{"bug", "list", "--product", "1"}},
		{"bug params", []string{"bug", "params", "--product", "1"}},
		{"bug create", []string{"bug", "create", "--product", "1", "--title", "Fix CSS"}},
		{"bug resolve-params", []string{"bug", "resolve-params", "--id", "1"}},
		{"bug resolve", []string{"bug", "resolve", "--id", "1", "--resolution", "fixed"}},
		{"bug delete", []string{"bug", "delete", "--id", "1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			flagOpts = config.Options{}
			RootCmd.SetArgs(append(tc.args, "-o", "json"))

			if err := RootCmd.Execute(); err != nil {
				t.Fatalf("command %q failed: %v", tc.name, err)
			}

			var resp map[string]any
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
				t.Fatalf("command %q invalid json output: %v, raw: %s", tc.name, err, buf.String())
			}
			if resp["ok"] != true {
				t.Fatalf("command %q returned ok != true: %s", tc.name, buf.String())
			}
		})
	}
}

func TestClassifyError_And_ValidationErrors(t *testing.T) {
	// 1. Test classifyError directly
	if code, cat := classifyError(nil); code != output.ExitCodeSuccess || cat != "none" {
		t.Errorf("unexpected success classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("session timeout please login")); code != output.ExitCodeAuth || cat != "auth" {
		t.Errorf("unexpected auth classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("--name is required")); code != output.ExitCodeValidation || cat != "validation" {
		t.Errorf("unexpected validation classification: code=%d, cat=%s", code, cat)
	}
	if code, cat := classifyError(fmt.Errorf("internal server error")); code != output.ExitCodeAPI || cat != "api" {
		t.Errorf("unexpected api classification: code=%d, cat=%s", code, cat)
	}

	// 2. Test unauthenticated command execution in clean environment
	t.Setenv("HOME", t.TempDir())
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"my", "task", "-o", "json"})
	err := RootCmd.Execute()
	if err == nil {
		t.Fatalf("expected unauthenticated error when running my task without login, got nil")
	}

	// 3. Test missing required flags
	validationCases := [][]string{
		{"product", "add"},
		{"project", "add"},
		{"dept", "add"},
		{"user", "add"},
		{"task", "list"},
		{"task", "params"},
		{"task", "create"},
		{"task", "finish"},
		{"task", "finish-params"},
		{"task", "delete"},
		{"bug", "list"},
		{"bug", "params"},
		{"bug", "create"},
		{"bug", "resolve"},
		{"bug", "resolve-params"},
		{"bug", "delete"},
		{"todo", "create"},
		{"todo", "start"},
		{"todo", "finish"},
		{"todo", "close"},
		{"todo", "delete"},
	}

	for _, args := range validationCases {
		buf.Reset()
		flagOpts = config.Options{}
		RootCmd.SetArgs(append(args, "-o", "json"))
		if err := RootCmd.Execute(); err == nil {
			t.Errorf("expected error for command %v without required flags, got nil", args)
		}
	}
}
