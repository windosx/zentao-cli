package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetTaskFlags resets package-level task flag variables to their declared
// defaults by reading DefValue from the relevant cobra commands.
func resetTaskFlags() {
	taskProjectID = ""
	taskID = ""
	taskStatus = taskListCmd.Flags().Lookup("status").DefValue
	taskOrderBy = taskListCmd.Flags().Lookup("order-by").DefValue
	taskName = ""
	taskType = taskCreateCmd.Flags().Lookup("type").DefValue
	taskPri = taskCreateCmd.Flags().Lookup("pri").DefValue
	taskEstimate = ""
	taskLeft = ""
	taskConsumed = ""
	taskAssignedTo = ""
	taskDesc = ""
	taskModuleID = taskCreateCmd.Flags().Lookup("module").DefValue
	taskStoryID = taskCreateCmd.Flags().Lookup("story").DefValue
	taskKeywords = ""
	taskMailto = ""
	taskDeadline = ""
	taskEstStarted = ""
	taskRealStarted = ""
	taskReal = taskFinishCmd.Flags().Lookup("real").DefValue
	taskComment = ""
	taskFinishedDate = ""
}

func TestTaskCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"task-sess-1","rand":"task-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case (m == "execution" || m == "project" || m == "task") && f == "task":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"101","name":"Task 1"}]`})
		case m == "task" && f == "view":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"id":"101","name":"Task 1"}`})
		case m == "task" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"projects":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "task" && f == "finish":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"task":{"id":"101"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "finished"})
			}
		case m == "task" && (f == "edit" || f == "start" || f == "pause" || f == "restart" || f == "close" || f == "cancel" || f == "activate" || f == "assign" || f == "delete"):
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
		// Validation errors (missing required flags)
		{"list missing project", []string{"task", "list"}, true},
		{"view missing id", []string{"task", "view"}, true},
		{"params missing project", []string{"task", "params"}, true},
		{"create missing project", []string{"task", "create", "--name", "Task 1"}, true},
		{"create missing name", []string{"task", "create", "--project", "109"}, true},
		{"edit missing id", []string{"task", "edit", "--name", "Updated"}, true},
		{"start missing id", []string{"task", "start"}, true},
		{"pause missing id", []string{"task", "pause"}, true},
		{"restart missing id", []string{"task", "restart"}, true},
		{"finish missing id", []string{"task", "finish"}, true},
		{"finish-params missing id", []string{"task", "finish-params"}, true},
		{"close missing id", []string{"task", "close"}, true},
		{"cancel missing id", []string{"task", "cancel"}, true},
		{"activate missing id", []string{"task", "activate"}, true},
		{"assign missing id", []string{"task", "assign", "--assigned-to", "dev"}, true},
		{"assign missing assigned-to", []string{"task", "assign", "--id", "101"}, true},
		{"delete missing id", []string{"task", "delete"}, true},

		// Execution tests for all 15 subcommands
		{"list success", []string{"task", "list", "--project", "109", "--status", "doing", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"task", "view", "--id", "101", "-o", "json"}, false},
		{"params success", []string{"task", "params", "--project", "109", "-o", "json"}, false},
		{"create full", []string{"task", "create", "--project", "109", "--name", "Dev API", "--type", "devel", "--pri", "2", "--estimate", "4.0", "--assigned-to", "admin", "--desc", "detail", "--keywords", "api,jwt", "--mailto", "dev1", "--deadline", "2026-09-01", "--est-started", "2026-08-30", "-o", "json"}, false},
		{"edit full", []string{"task", "edit", "--id", "101", "--name", "Dev API V2", "--keywords", "api,v2", "--comment", "update", "-o", "json"}, false},
		{"start full", []string{"task", "start", "--id", "101", "--real-started", "2026-08-28 09:00:00", "-o", "json"}, false},
		{"pause success", []string{"task", "pause", "--id", "101", "--comment", "pause", "-o", "json"}, false},
		{"restart success", []string{"task", "restart", "--id", "101", "--comment", "restart", "-o", "json"}, false},
		{"finish-params success", []string{"task", "finish-params", "--id", "101", "-o", "json"}, false},
		{"finish success", []string{"task", "finish", "--id", "101", "--real", "3.0", "--comment", "done", "-o", "json"}, false},
		{"close success", []string{"task", "close", "--id", "101", "-o", "json"}, false},
		{"cancel success", []string{"task", "cancel", "--id", "101", "--comment", "cancel", "-o", "json"}, false},
		{"activate success", []string{"task", "activate", "--id", "101", "-o", "json"}, false},
		{"assign success", []string{"task", "assign", "--id", "101", "--assigned-to", "admin", "-o", "json"}, false},
		{"delete success", []string{"task", "delete", "--id", "101", "--project", "109", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTaskFlags()
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
