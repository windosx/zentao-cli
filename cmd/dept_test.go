package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetDeptFlags resets package-level dept flag variables. All dept flags
// default to empty strings or nil, so no DefValue lookup is needed.
func resetDeptFlags() {
	deptID = ""
	deptParentID = ""
	deptName = ""
	deptManager = ""
	deptNames = nil
}

func TestDeptCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"dept-sess-1","rand":"dept-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "dept" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"研发"}]`})
		case m == "dept" && f == "manageChild":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
		case m == "dept" && (f == "edit" || f == "delete"):
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
		{"create missing name", []string{"dept", "create", "--parent", "1"}, true},
		{"edit missing id", []string{"dept", "edit", "--name", "Updated"}, true},
		{"delete missing id", []string{"dept", "delete"}, true},

		// Execution tests for all 4 subcommands
		{"list with pagination", []string{"dept", "list", "--parent", "0", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"create full", []string{"dept", "create", "--parent", "1", "--name", "Frontend", "--name", "Backend", "-o", "json"}, false},
		{"edit full", []string{"dept", "edit", "--id", "1", "--name", "R&D", "--manager", "admin", "-o", "json"}, false},
		{"delete success", []string{"dept", "delete", "--id", "1", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetDeptFlags()
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
