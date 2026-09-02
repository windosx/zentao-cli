package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

func resetTrashFlags() {
	trashType = trashListCmd.Flags().Lookup("type").DefValue
	trashOrderBy = trashListCmd.Flags().Lookup("order-by").DefValue
	trashActionID = ""
}

func TestTrashCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"trash-sess-1","rand":"trash-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "action" && f == "trash":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"trashes":[{"id":123,"objectType":"task","objectID":101},{"id":124,"objectType":"bug","objectID":202}]}`,
			})
		case m == "action" && (f == "undelete" || f == "hideOne" || f == "hideAll"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"result":  "success",
				"message": "ok",
			})
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
		{"restore missing action-id", []string{"trash", "restore"}, true},
		{"hide-one missing action-id", []string{"trash", "hide-one"}, true},

		// Execution tests for trash subcommands
		{"list default", []string{"trash", "list", "-o", "json"}, false},
		{"list with type and order", []string{"trash", "list", "--type", "task", "--order-by", "date_desc", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"restore success", []string{"trash", "restore", "--action-id", "123", "-o", "json"}, false},
		{"hide-one success", []string{"trash", "hide-one", "--action-id", "123", "-o", "json"}, false},
		{"hide-all success", []string{"trash", "hide-all", "-o", "json"}, false},

		// Object-level restore commands using trash search
		{"task restore missing id", []string{"task", "restore"}, true},
		{"task restore success", []string{"task", "restore", "--id", "101", "-o", "json"}, false},
		{"bug restore missing id", []string{"bug", "restore"}, true},
		{"bug restore success", []string{"bug", "restore", "--id", "202", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTrashFlags()
			resetTaskFlags()
			resetBugFlags()
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
