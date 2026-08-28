package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/windosx/zentao-cli/internal/config"
)

// resetProductFlags resets package-level product flag variables to their
// declared defaults, read from the cobra command flag definitions.
func resetProductFlags() {
	productID = ""
	productStatus = productListCmd.Flags().Lookup("status").DefValue
	productOrderBy = productListCmd.Flags().Lookup("order-by").DefValue
	productLineID = productListCmd.Flags().Lookup("line").DefValue
	productProgramID = productListCmd.Flags().Lookup("program").DefValue
	productName = ""
	productCode = ""
	productType = productCreateCmd.Flags().Lookup("type").DefValue
	productPO = ""
	productQD = ""
	productRD = ""
	productACL = productCreateCmd.Flags().Lookup("acl").DefValue
	productWhitelist = ""
	productDesc = ""
	productComment = ""
}

func TestProductCommands_ValidationAndExecution(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ZENTAO_NO_KEYRING", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "api" && f == "getSessionID":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"prod-sess-1","rand":"prod-rand-1"}`,
			})
		case m == "user" && f == "login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"user":   map[string]any{"id": "1", "account": "admin"},
			})
		case m == "product" && (f == "browse" || f == "all" || f == "view"):
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"8","name":"Product 1"}]`})
		case m == "product" && f == "create":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `{"products":[]}`})
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
			}
		case m == "product" && (f == "edit" || f == "close" || f == "activate" || f == "delete"):
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
		{"view missing id", []string{"product", "view"}, true},
		{"create missing name", []string{"product", "create", "--code", "app"}, true},
		{"create missing code", []string{"product", "create", "--name", "App"}, true},
		{"edit missing id", []string{"product", "edit", "--name", "Updated"}, true},
		{"close missing id", []string{"product", "close"}, true},
		{"activate missing id", []string{"product", "activate"}, true},
		{"delete missing id", []string{"product", "delete"}, true},

		// Execution tests for all 8 subcommands
		{"list with pagination", []string{"product", "list", "--status", "noclosed", "--limit", "10", "--page", "1", "-o", "json"}, false},
		{"view success", []string{"product", "view", "--id", "8", "-o", "json"}, false},
		{"params success", []string{"product", "params", "--program", "0", "-o", "json"}, false},
		{"create full", []string{"product", "create", "--name", "Mobile App", "--code", "app", "--line", "1", "--program", "0", "--type", "normal", "--po", "admin", "--qd", "admin", "--rd", "admin", "--acl", "open", "--status", "normal", "--desc", "product desc", "-o", "json"}, false},
		{"edit full", []string{"product", "edit", "--id", "8", "--name", "Mobile App Pro", "--comment", "updating", "-o", "json"}, false},
		{"close success", []string{"product", "close", "--id", "8", "--comment", "closed", "-o", "json"}, false},
		{"activate success", []string{"product", "activate", "--id", "8", "--comment", "activated", "-o", "json"}, false},
		{"delete success", []string{"product", "delete", "--id", "8", "-o", "json"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetProductFlags()
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
