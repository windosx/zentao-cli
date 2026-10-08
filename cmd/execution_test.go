package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func resetExecutionFlags() {
	executionID = ""
	executionProjectID = ""
	executionStatus = executionListCmd.Flags().Lookup("status").DefValue
	executionOrderBy = executionListCmd.Flags().Lookup("order-by").DefValue
	executionName = ""
	executionCode = ""
	executionBegin = ""
	executionEnd = ""
	executionDays = ""
	executionTeam = ""
	executionType = executionCreateCmd.Flags().Lookup("type").DefValue
	executionPri = executionCreateCmd.Flags().Lookup("pri").DefValue
	executionPM = ""
	executionPO = ""
	executionQD = ""
	executionRD = ""
	executionACL = executionCreateCmd.Flags().Lookup("acl").DefValue
	executionWhitelist = ""
	executionDesc = ""
	executionComment = ""
}

func TestExecutionCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"exec-sess-1","rand":"exec-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case (m == "project" && f == "execution") || (m == "execution" && (f == "browse" || f == "all" || f == "view")):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"501","name":"Sprint 1"}]`})
		case m == "execution" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"projects":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "execution" && (f == "edit" || f == "start" || f == "suspend" || f == "activate" || f == "close" || f == "delete"):
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "ok"})
		case m == "execution" && f == "task":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"101","name":"T1"}]`})
		case m == "execution" && f == "story":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"201","title":"S1"}]`})
		case m == "execution" && f == "bug":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"301","title":"B1"}]`})
		case m == "action" && f == "trash":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"trashes":[{"id":"999","objectType":"execution","objectID":"501"}]}`,
			})
		case m == "action" && f == "undelete":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"result":"success"}`})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{}`})
		}
	}))
	defer server.Close()

	// Initial auth login
	flagOpts = config.Options{}
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"auth", "login", "-u", server.URL, "-a", "admin", "-p", "123456", "-o", "json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("auth login failed: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		// Validation tests
		{"view missing id", []string{"execution", "view"}, true},
		{"create missing name", []string{"execution", "create", "--code", "s1"}, true},
		{"create missing code", []string{"execution", "create", "--name", "Sprint 1"}, true},
		{"edit missing id", []string{"execution", "edit", "--name", "Updated"}, true},
		{"start missing id", []string{"execution", "start"}, true},
		{"suspend missing id", []string{"execution", "suspend"}, true},
		{"activate missing id", []string{"execution", "activate"}, true},
		{"close missing id", []string{"execution", "close"}, true},
		{"delete missing id", []string{"execution", "delete"}, true},
		{"restore missing id", []string{"execution", "restore"}, true},
		{"task missing id", []string{"execution", "task"}, true},
		{"story missing id", []string{"execution", "story"}, true},
		{"bug missing id", []string{"execution", "bug"}, true},

		// Execution tests
		{"list with pagination", []string{"execution", "list", "--status", "doing", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"execution", "view", "--id", "501", "-o", "json"}, false},
		{"params success", []string{"execution", "params", "--project", "1", "-o", "json"}, false},
		{"create full", []string{"execution", "create", "--project", "1", "--name", "Sprint 1", "--code", "s1", "--begin", "2026-10-01", "--end", "2026-10-15", "--days", "10", "--team", "Alpha", "--type", "sprint", "--pri", "1", "--pm", "admin", "--po", "admin", "--qd", "admin", "--rd", "admin", "--desc", "Sprint description", "-o", "json"}, false},
		{"edit full", []string{"execution", "edit", "--id", "501", "--name", "Sprint 1 Renamed", "--comment", "updating name", "-o", "json"}, false},
		{"start success", []string{"execution", "start", "--id", "501", "--comment", "started", "-o", "json"}, false},
		{"suspend success", []string{"execution", "suspend", "--id", "501", "--comment", "suspended", "-o", "json"}, false},
		{"activate success", []string{"execution", "activate", "--id", "501", "--comment", "activated", "-o", "json"}, false},
		{"close success", []string{"execution", "close", "--id", "501", "--comment", "closed", "-o", "json"}, false},
		{"delete success", []string{"execution", "delete", "--id", "501", "-o", "json"}, false},
		{"restore success", []string{"execution", "restore", "--id", "501", "-o", "json"}, false},
		{"task association", []string{"execution", "task", "--id", "501", "-o", "json"}, false},
		{"story association", []string{"execution", "story", "--id", "501", "-o", "json"}, false},
		{"bug association", []string{"execution", "bug", "--id", "501", "-o", "json"}, false},

		// Alias tests
		{"alias iter list", []string{"iter", "list", "-o", "json"}, false},
		{"alias sprint view", []string{"sprint", "view", "--id", "501", "-o", "json"}, false},
		{"alias iteration list", []string{"iteration", "list", "-o", "json"}, false},
		{"alias exec list", []string{"exec", "list", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetExecutionFlags()
			buf.Reset()
			flagOpts = config.Options{}
			RootCmd.SetArgs(tt.args)

			err := RootCmd.Execute()
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected validation error for %v, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error executing %v: %v\nOutput: %s", tt.args, err, buf.String())
			}

			var resp map[string]any
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
				t.Fatalf("expected ok=true response for %v, got: %s", tt.args, buf.String())
			}
		})
	}
}
