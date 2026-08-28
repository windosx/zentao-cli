package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetTodoFlags resets package-level todo flag variables to their declared
// defaults, read from the cobra command flag definitions.
func resetTodoFlags() {
	todoType = todoListCmd.Flags().Lookup("type").DefValue
	todoStatus = todoListCmd.Flags().Lookup("status").DefValue
	todoOrderBy = todoListCmd.Flags().Lookup("order-by").DefValue
	todoID = ""
	todoName = ""
	todoDate = ""
	todoBegin = ""
	todoEnd = ""
	todoPri = todoCreateCmd.Flags().Lookup("pri").DefValue
	todoDesc = ""
	todoIDValue = todoCreateCmd.Flags().Lookup("idvalue").DefValue
	todoPrivate = todoCreateCmd.Flags().Lookup("private").DefValue
	todoAssignedTo = ""
}

func TestTodoCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"todo-sess-1","rand":"todo-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "todo" && (f == "browse" || f == "view"):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"3","name":"Todo 1"}]`})
		case m == "todo" && f == "create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
		case m == "todo" && (f == "edit" || f == "start" || f == "finish" || f == "close" || f == "activate" || f == "assign" || f == "delete"):
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
		{"view missing id", []string{"todo", "view"}, true},
		{"create missing name", []string{"todo", "create", "--pri", "1"}, true},
		{"edit missing id", []string{"todo", "edit", "--name", "Updated"}, true},
		{"start missing id", []string{"todo", "start"}, true},
		{"finish missing id", []string{"todo", "finish"}, true},
		{"close missing id", []string{"todo", "close"}, true},
		{"activate missing id", []string{"todo", "activate"}, true},
		{"assign missing id", []string{"todo", "assign", "--assigned-to", "admin"}, true},
		{"assign missing assigned-to", []string{"todo", "assign", "--id", "3"}, true},
		{"delete missing id", []string{"todo", "delete"}, true},

		// Execution tests for all 10 subcommands
		{"list with pagination", []string{"todo", "list", "--type", "today", "--status", "wait", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"todo", "view", "--id", "3", "-o", "json"}, false},
		{"create full", []string{"todo", "create", "--name", "Review Code", "--date", "2026-08-28", "--begin", "0900", "--end", "1800", "--type", "custom", "--pri", "2", "--desc", "code review details", "--private", "0", "-o", "json"}, false},
		{"edit full", []string{"todo", "edit", "--id", "3", "--name", "Review Code V2", "--status", "doing", "-o", "json"}, false},
		{"start success", []string{"todo", "start", "--id", "3", "-o", "json"}, false},
		{"finish success", []string{"todo", "finish", "--id", "3", "-o", "json"}, false},
		{"close success", []string{"todo", "close", "--id", "3", "-o", "json"}, false},
		{"activate success", []string{"todo", "activate", "--id", "3", "-o", "json"}, false},
		{"assign success", []string{"todo", "assign", "--id", "3", "--assigned-to", "admin", "-o", "json"}, false},
		{"delete success", []string{"todo", "delete", "--id", "3", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTodoFlags()
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
