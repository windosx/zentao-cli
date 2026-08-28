package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetProjectFlags resets package-level project flag variables to their
// declared defaults, read from the cobra command flag definitions.
func resetProjectFlags() {
	projectID = ""
	projectStatus = projectListCmd.Flags().Lookup("status").DefValue
	projectOrderBy = projectListCmd.Flags().Lookup("order-by").DefValue
	projectProductID = ""
	projectProgramID = projectListCmd.Flags().Lookup("program").DefValue
	projectName = ""
	projectCode = ""
	projectBegin = ""
	projectEnd = ""
	projectDays = ""
	projectTeam = ""
	projectType = projectCreateCmd.Flags().Lookup("type").DefValue
	projectPM = ""
	projectPO = ""
	projectQD = ""
	projectRD = ""
	projectPri = projectCreateCmd.Flags().Lookup("pri").DefValue
	projectACL = projectCreateCmd.Flags().Lookup("acl").DefValue
	projectWhitelist = ""
	projectDesc = ""
	projectComment = ""
}

func TestProjectCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"proj-sess-1","rand":"proj-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case (m == "project" || m == "program" || m == "execution") && (f == "browse" || f == "all" || f == "view"):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"109","name":"Sprint 2"}]`})
		case m == "project" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"allProducts":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "project" && (f == "edit" || f == "start" || f == "suspend" || f == "activate" || f == "close" || f == "delete"):
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "ok"})
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
		{"view missing id", []string{"project", "view"}, true},
		{"create missing name", []string{"project", "create", "--code", "s2"}, true},
		{"create missing code", []string{"project", "create", "--name", "Sprint 2"}, true},
		{"edit missing id", []string{"project", "edit", "--name", "Updated"}, true},
		{"start missing id", []string{"project", "start"}, true},
		{"suspend missing id", []string{"project", "suspend"}, true},
		{"activate missing id", []string{"project", "activate"}, true},
		{"close missing id", []string{"project", "close"}, true},
		{"delete missing id", []string{"project", "delete"}, true},

		// Execution tests for all 10 subcommands
		{"list with pagination", []string{"project", "list", "--status", "doing", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"project", "view", "--id", "109", "-o", "json"}, false},
		{"params success", []string{"project", "params", "--program", "0", "-o", "json"}, false},
		{"create full", []string{"project", "create", "--name", "Sprint 2", "--code", "s2", "--program", "0", "--begin", "2026-09-01", "--end", "2026-09-15", "--days", "10", "--team", "Team Alpha", "--type", "sprint", "--pri", "2", "--pm", "admin", "--po", "admin", "--desc", "iteration", "-o", "json"}, false},
		{"edit full", []string{"project", "edit", "--id", "109", "--name", "Sprint 2 Adjusted", "--comment", "updating", "-o", "json"}, false},
		{"start success", []string{"project", "start", "--id", "109", "--comment", "started", "-o", "json"}, false},
		{"suspend success", []string{"project", "suspend", "--id", "109", "--comment", "suspended", "-o", "json"}, false},
		{"activate success", []string{"project", "activate", "--id", "109", "--comment", "activated", "-o", "json"}, false},
		{"close success", []string{"project", "close", "--id", "109", "--comment", "closed", "-o", "json"}, false},
		{"delete success", []string{"project", "delete", "--id", "109", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetProjectFlags()
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
