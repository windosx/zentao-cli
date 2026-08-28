package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetMyFlags resets package-level my-command flag variables to their
// declared defaults, read from the cobra command flag definitions.
func resetMyFlags() {
	myTaskType = myTaskCmd.Flags().Lookup("type").DefValue
	myTaskOrderBy = myTaskCmd.Flags().Lookup("order-by").DefValue
	myBugType = myBugCmd.Flags().Lookup("type").DefValue
	myBugOrderBy = myBugCmd.Flags().Lookup("order-by").DefValue
	myStoryType = myStoryCmd.Flags().Lookup("type").DefValue
	myStoryOrderBy = myStoryCmd.Flags().Lookup("order-by").DefValue
	myTodoType = myTodoCmd.Flags().Lookup("type").DefValue
	myTodoStatus = myTodoCmd.Flags().Lookup("status").DefValue
	myProjectType = myProjectCmd.Flags().Lookup("status").DefValue
	myProjectOrderBy = myProjectCmd.Flags().Lookup("order-by").DefValue
	myDynamicType = myDynamicCmd.Flags().Lookup("type").DefValue
}

func TestMyCommands_ExecutionAndPagination(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"my-sess-1","rand":"my-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "my" && (f == "task" || f == "bug" || f == "todo" || f == "story" || f == "project" || f == "dynamic"):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[]`})
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
		name string
		args []string
	}{
		// 1. my task with various types
		{"my task default", []string{"my", "task", "-o", "json"}},
		{"my task finishedBy", []string{"my", "task", "--type", "finishedBy", "-o", "json"}},
		{"my task openedBy", []string{"my", "task", "--type", "openedBy", "-o", "json"}},
		{"my task undone", []string{"my", "task", "--type", "undone", "--limit", "20", "--page", "1", "-o", "json"}},

		// 2. my bug with types
		{"my bug default", []string{"my", "bug", "-o", "json"}},
		{"my bug openedBy", []string{"my", "bug", "--type", "openedBy", "-o", "json"}},
		{"my bug resolvedBy", []string{"my", "bug", "--type", "resolvedBy", "-o", "json"}},

		// 3. my todo with types
		{"my todo default", []string{"my", "todo", "-o", "json"}},
		{"my todo today", []string{"my", "todo", "--type", "today", "-o", "json"}},
		{"my todo thisWeek", []string{"my", "todo", "--type", "thisWeek", "-o", "json"}},
		{"my todo before", []string{"my", "todo", "--type", "before", "-o", "json"}},

		// 4. my story
		{"my story default", []string{"my", "story", "-o", "json"}},
		{"my story assignedTo", []string{"my", "story", "--type", "assignedTo", "-o", "json"}},

		// 5. my project
		{"my project doing", []string{"my", "project", "--status", "doing", "-o", "json"}},
		{"my project all", []string{"my", "project", "--status", "all", "-o", "json"}},

		// 6. my dynamic
		{"my dynamic today", []string{"my", "dynamic", "--type", "today", "-o", "json"}},
		{"my dynamic thisWeek", []string{"my", "dynamic", "--type", "thisWeek", "-o", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMyFlags()
			buf.Reset()
			flagOpts = config.Options{}
			RootCmd.SetArgs(tt.args)

			if err := RootCmd.Execute(); err != nil {
				t.Fatalf("unexpected error executing %v: %v\nOutput: %s", tt.args, err, buf.String())
			}

			var resp map[string]any
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil || resp["ok"] != true {
				t.Fatalf("expected ok=true response for %v, got: %s", tt.args, buf.String())
			}
		})
	}
}
