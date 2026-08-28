package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetBugFlags resets package-level bug flag variables to their declared
// defaults by reading the DefValue from the relevant cobra commands. This
// keeps the reset in sync with the production flag definitions automatically.
func resetBugFlags() {
	bugProductID = ""
	bugID = ""
	bugBranch = bugListCmd.Flags().Lookup("branch").DefValue
	bugBrowseType = bugListCmd.Flags().Lookup("browse-type").DefValue
	bugOrderBy = bugListCmd.Flags().Lookup("order-by").DefValue
	bugTitle = ""
	bugSeverity = bugCreateCmd.Flags().Lookup("severity").DefValue
	bugPri = bugCreateCmd.Flags().Lookup("pri").DefValue
	bugType = bugCreateCmd.Flags().Lookup("type").DefValue
	bugAssignedTo = ""
	bugSteps = ""
	bugOpenedBuild = bugCreateCmd.Flags().Lookup("opened-build").DefValue
	bugProjectID = bugCreateCmd.Flags().Lookup("project").DefValue
	bugStoryID = bugCreateCmd.Flags().Lookup("story").DefValue
	bugModuleID = bugCreateCmd.Flags().Lookup("module").DefValue
	bugKeywords = ""
	bugMailto = ""
	bugOS = ""
	bugBrowser = ""
	bugHardware = ""
	bugFound = ""
	bugDeadline = ""
	bugDuplicateBug = ""
	bugStatus = ""
	bugResolution = bugResolveCmd.Flags().Lookup("resolution").DefValue
	bugResolvedBuild = bugResolveCmd.Flags().Lookup("resolved-build").DefValue
	bugResolvedDate = ""
	bugComment = ""
}

func TestBugCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"bug-sess-1","rand":"bug-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "bug" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"2862","title":"Bug 1"}]`})
		case m == "bug" && f == "view":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"id":"2862","title":"Bug 1"}`})
		case m == "bug" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"types":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "bug created"})
			}
		case m == "bug" && f == "resolve":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"bug":{"id":"2862"}}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "bug resolved"})
			}
		case m == "bug" && (f == "edit" || f == "close" || f == "activate" || f == "assign" || f == "confirm" || f == "delete"):
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
		{"list missing product", []string{"bug", "list"}, true},
		{"view missing id", []string{"bug", "view"}, true},
		{"params missing product", []string{"bug", "params"}, true},
		{"create missing product", []string{"bug", "create", "--title", "Crash"}, true},
		{"create missing title", []string{"bug", "create", "--product", "8"}, true},
		{"edit missing id", []string{"bug", "edit", "--title", "Updated"}, true},
		{"resolve-params missing id", []string{"bug", "resolve-params"}, true},
		{"resolve missing id", []string{"bug", "resolve"}, true},
		{"close missing id", []string{"bug", "close"}, true},
		{"activate missing id", []string{"bug", "activate"}, true},
		{"assign missing id", []string{"bug", "assign", "--assigned-to", "dev"}, true},
		{"assign missing assigned-to", []string{"bug", "assign", "--id", "2862"}, true},
		{"confirm missing id", []string{"bug", "confirm"}, true},
		{"delete missing id", []string{"bug", "delete"}, true},

		// Execution tests for all 12 subcommands
		{"list with pagination", []string{"bug", "list", "--product", "8", "--browse-type", "unclosed", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"bug", "view", "--id", "2862", "-o", "json"}, false},
		{"params success", []string{"bug", "params", "--product", "8", "-o", "json"}, false},
		{"create full", []string{"bug", "create", "--product", "8", "--title", "Login Crash", "--severity", "2", "--pri", "2", "--type", "codeerror", "--opened-build", "trunk", "--assigned-to", "admin", "--steps", "click login", "--project", "109", "--story", "55", "--module", "1", "--keywords", "crash,safari", "--mailto", "dev1", "--os", "mac", "--browser", "safari", "--deadline", "2026-09-01", "-o", "json"}, false},
		{"edit full", []string{"bug", "edit", "--id", "2862", "--title", "Login Crash V2", "--keywords", "crash,v2", "--comment", "updating", "-o", "json"}, false},
		{"resolve-params success", []string{"bug", "resolve-params", "--id", "2862", "-o", "json"}, false},
		{"resolve full", []string{"bug", "resolve", "--id", "2862", "--resolution", "fixed", "--resolved-build", "1.0.1", "--resolved-date", "2026-08-28", "--comment", "fixed in pr 1", "-o", "json"}, false},
		{"close success", []string{"bug", "close", "--id", "2862", "--comment", "verified and closed", "-o", "json"}, false},
		{"activate success", []string{"bug", "activate", "--id", "2862", "--opened-build", "trunk", "--assigned-to", "admin", "--comment", "reproduced", "-o", "json"}, false},
		{"assign success", []string{"bug", "assign", "--id", "2862", "--assigned-to", "admin", "--comment", "reassign", "-o", "json"}, false},
		{"confirm success", []string{"bug", "confirm", "--id", "2862", "--pri", "1", "--comment", "confirmed", "-o", "json"}, false},
		{"delete success", []string{"bug", "delete", "--id", "2862", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
