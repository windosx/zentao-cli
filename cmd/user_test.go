package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetUserFlags resets package-level user flag variables to their declared
// defaults, read from the cobra command flag definitions.
func resetUserFlags() {
	userID = ""
	userDeptID = userListCmd.Flags().Lookup("dept").DefValue
	userType = userListCmd.Flags().Lookup("type").DefValue
	userOrderBy = userListCmd.Flags().Lookup("order-by").DefValue
	newUsername = ""
	newUserPassword = ""
	userRealname = ""
	userRole = userCreateCmd.Flags().Lookup("role").DefValue
	userEmail = ""
	userGender = userCreateCmd.Flags().Lookup("gender").DefValue
	userMobile = ""
	userPhone = ""
	userQQ = ""
}

func TestUserCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"user-sess-1","rand":"user-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "company" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"12","account":"tom"}]`})
		case m == "user" && f == "view":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"id":"12","account":"tom"}`})
		case m == "user" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"depts":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "user" && (f == "edit" || f == "delete"):
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
		{"view missing id", []string{"user", "view"}, true},
		{"create missing username", []string{"user", "create", "--user-password", "p", "--realname", "Tom"}, true},
		{"create missing password", []string{"user", "create", "--username", "tom", "--realname", "Tom"}, true},
		{"create missing realname", []string{"user", "create", "--username", "tom", "--user-password", "p"}, true},
		{"edit missing id", []string{"user", "edit", "--realname", "Updated"}, true},
		{"delete missing id", []string{"user", "delete"}, true},

		// Execution tests for all 6 subcommands
		{"list with pagination", []string{"user", "list", "--dept", "1", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"user", "view", "--id", "12", "-o", "json"}, false},
		{"params success", []string{"user", "params", "--dept", "1", "-o", "json"}, false},
		{"create full", []string{"user", "create", "--username", "tom", "--user-password", "pwd123", "--realname", "Tom Hanks", "--role", "dev", "--gender", "m", "--email", "tom@example.com", "--mobile", "13800000000", "--dept", "1", "-o", "json"}, false},
		{"edit full", []string{"user", "edit", "--id", "12", "--realname", "Thomas", "-o", "json"}, false},
		{"delete success", []string{"user", "delete", "--id", "12", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetUserFlags()
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
