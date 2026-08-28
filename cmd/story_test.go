package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetStoryFlags resets package-level story flag variables to their declared
// defaults, read from the cobra command flag definitions.
func resetStoryFlags() {
	storyProductID = ""
	storyID = ""
	storyBranch = storyListCmd.Flags().Lookup("branch").DefValue
	storyBrowseType = storyListCmd.Flags().Lookup("browse-type").DefValue
	storyType = storyListCmd.Flags().Lookup("type").DefValue
	storyOrderBy = storyListCmd.Flags().Lookup("order-by").DefValue
	storyTitle = ""
	storyPri = storyCreateCmd.Flags().Lookup("pri").DefValue
	storyEstimate = ""
	storyAssignedTo = ""
	storyModuleID = storyCreateCmd.Flags().Lookup("module").DefValue
	storyPlanID = storyCreateCmd.Flags().Lookup("plan").DefValue
	storySource = ""
	storySourceNote = ""
	storyKeywords = ""
	storyMailto = ""
	storySpec = ""
	storyVerify = ""
	storyStatus = ""
	storyResult = storyReviewCmd.Flags().Lookup("result").DefValue
	storyReason = storyCloseCmd.Flags().Lookup("reason").DefValue
	storyComment = ""
}

func TestStoryCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"story-sess-1","rand":"story-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "story" && (f == "browse" || f == "view"):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"id":"55","title":"Story 1"}`})
		case m == "story" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"products":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "story" && (f == "edit" || f == "review" || f == "change" || f == "close" || f == "activate" || f == "assign" || f == "delete"):
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
		{"list missing product", []string{"story", "list"}, true},
		{"view missing id", []string{"story", "view"}, true},
		{"params missing product", []string{"story", "params"}, true},
		{"create missing product", []string{"story", "create", "--title", "Feature"}, true},
		{"create missing title", []string{"story", "create", "--product", "8"}, true},
		{"edit missing id", []string{"story", "edit", "--title", "Updated"}, true},
		{"review missing id", []string{"story", "review", "--result", "pass"}, true},
		{"change missing id", []string{"story", "change", "--spec", "new spec"}, true},
		{"close missing id", []string{"story", "close"}, true},
		{"activate missing id", []string{"story", "activate"}, true},
		{"assign missing id", []string{"story", "assign", "--assigned-to", "dev"}, true},
		{"assign missing assigned-to", []string{"story", "assign", "--id", "55"}, true},
		{"delete missing id", []string{"story", "delete"}, true},

		// Execution tests for all 11 subcommands
		{"list with pagination", []string{"story", "list", "--product", "8", "--branch", "all", "--browse-type", "unclosed", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"story", "view", "--id", "55", "-o", "json"}, false},
		{"params success", []string{"story", "params", "--product", "8", "-o", "json"}, false},
		{"create full", []string{"story", "create", "--product", "8", "--title", "Wechat Pay", "--type", "story", "--pri", "2", "--estimate", "8.0", "--assigned-to", "admin", "--module", "1", "--plan", "2", "--source", "po", "--source-note", "v1", "--keywords", "pay,wechat", "--mailto", "dev1", "--spec", "user can pay", "--verify", "unit test pass", "-o", "json"}, false},
		{"edit full", []string{"story", "edit", "--id", "55", "--title", "Wechat Pay V2", "--keywords", "pay,v2", "--comment", "updating", "-o", "json"}, false},
		{"review success", []string{"story", "review", "--id", "55", "--result", "pass", "--comment", "reviewed", "-o", "json"}, false},
		{"change success", []string{"story", "change", "--id", "55", "--spec", "updated spec", "--verify", "updated verify", "--comment", "spec changed", "-o", "json"}, false},
		{"close success", []string{"story", "close", "--id", "55", "--reason", "done", "--comment", "finished", "-o", "json"}, false},
		{"activate success", []string{"story", "activate", "--id", "55", "--assigned-to", "admin", "--comment", "reopened", "-o", "json"}, false},
		{"assign success", []string{"story", "assign", "--id", "55", "--assigned-to", "admin", "--comment", "assigned", "-o", "json"}, false},
		{"delete success", []string{"story", "delete", "--id", "55", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStoryFlags()
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
