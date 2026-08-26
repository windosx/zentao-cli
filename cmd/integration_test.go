//go:build integration

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/windosx/zentao-cli/internal/config"
	"github.com/windosx/zentao-cli/pkg/zentao"
)

// getTestClient returns an authenticated client for integration testing against a live ZenTao server.
func getTestClient(t *testing.T) *zentao.Client {
	url := os.Getenv("ZENTAO_TEST_URL")
	account := os.Getenv("ZENTAO_TEST_ACCOUNT")
	password := os.Getenv("ZENTAO_TEST_PASSWORD")

	// If env vars not set, fallback to ~/.config/zentao/profiles.json
	if url == "" || account == "" || password == "" {
		profile, err := config.GetActiveProfile("")
		if err != nil || profile == nil {
			t.Skip("Skipping integration test: no active profile in ~/.config/zentao/profiles.json and no ZENTAO_TEST_* env vars")
		}
		url = profile.URL
		account = profile.Account
		password = profile.Password
	}

	if url == "" || account == "" {
		t.Skip("Skipping integration test: missing URL or Account")
	}

	c := zentao.New(zentao.Config{
		URL:        url,
		Account:    account,
		Password:   password,
		AccessMode: zentao.AccessModeGET,
		Timeout:    30 * time.Second,
	})

	ctx := context.Background()
	if err := c.Login(ctx); err != nil {
		t.Fatalf("Integration test login failed: %v", err)
	}

	return c
}

func execCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(errBuf)
	flagOpts = config.Options{}
	RootCmd.SetArgs(args)

	err := RootCmd.Execute()
	if err != nil {
		return errBuf.String(), err
	}
	return buf.String(), nil
}

// TestIntegration_FullLifecycle verifies all CLI commands against a live ZenTao server
// and strictly cleans up all test-generated records using defer / t.Cleanup.
func TestIntegration_FullLifecycle(t *testing.T) {
	c := getTestClient(t)
	ctx := context.Background()

	// 1. Auth Status & List
	t.Run("Auth_Commands", func(t *testing.T) {
		out, err := execCmd("auth", "status", "-o", "json")
		if err != nil {
			t.Fatalf("auth status failed: %v, out: %s", err, out)
		}
		if !strings.Contains(out, `"status": "authenticated"`) {
			t.Errorf("unexpected auth status output: %s", out)
		}

		outList, errList := execCmd("auth", "list", "-o", "json")
		if errList != nil {
			t.Fatalf("auth list failed: %v, out: %s", errList, outList)
		}
		if !strings.Contains(outList, `"profiles"`) {
			t.Errorf("unexpected auth list output: %s", outList)
		}
	})

	// 2. Personal Workbench (My Task, Bug, Todo, Story, Project, Dynamic) with multiple filter options
	t.Run("My_Workbench_Filters", func(t *testing.T) {
		myFilters := [][]string{
			{"my", "task", "--type", "assignedTo", "-o", "json"},
			{"my", "task", "--type", "finishedBy", "-o", "table"},
			{"my", "task", "--type", "openedBy", "-o", "text"},
			{"my", "bug", "--type", "assignedTo", "-o", "json"},
			{"my", "bug", "--type", "openedBy", "-o", "table"},
			{"my", "todo", "--type", "all", "--status", "all", "-o", "table"},
			{"my", "todo", "--type", "today", "-o", "json"},
			{"my", "todo", "--type", "thisWeek", "-o", "json"},
			{"my", "story", "--type", "assignedTo", "-o", "table"},
			{"my", "project", "--status", "doing", "-o", "table"},
			{"my", "project", "--status", "all", "-o", "json"},
			{"my", "dynamic", "--type", "today", "-o", "text"},
		}

		for _, args := range myFilters {
			name := strings.Join(args, "_")
			t.Run(name, func(t *testing.T) {
				out, err := execCmd(args...)
				if err != nil {
					t.Fatalf("command %v failed: %v, out: %s", args, err, out)
				}
				if len(out) == 0 {
					t.Errorf("expected non-empty output for %v", args)
				}
			})
		}
	})

	// 3. Product & Project List with Filters
	t.Run("Product_And_Project_Filters", func(t *testing.T) {
		queries := [][]string{
			{"product", "list", "--status", "normal", "-o", "table"},
			{"product", "list", "--status", "all", "-o", "json"},
			{"project", "list", "--status", "doing", "-o", "table"},
			{"project", "list", "--status", "undone", "-o", "json"},
			{"project", "list", "--status", "all", "-o", "table"},
			{"dept", "list", "-o", "table"},
			{"user", "list", "-o", "table"},
		}

		for _, args := range queries {
			name := strings.Join(args, "_")
			t.Run(name, func(t *testing.T) {
				out, err := execCmd(args...)
				if err != nil {
					t.Fatalf("command %v failed: %v, out: %s", args, err, out)
				}
			})
		}
	})

	// 4. Todo Lifecycle: Create -> Start -> Finish -> Close -> Cleanup (Delete)
	t.Run("Todo_Lifecycle_ZeroArtifact", func(t *testing.T) {
		uniqueName := fmt.Sprintf("【CI测试待办-%d】", time.Now().UnixNano())
		today := time.Now().Format("2006-01-02")

		// 4.1 Create
		out, err := execCmd("todo", "create", "--name", uniqueName, "--date", today, "--pri", "2", "-o", "json")
		if err != nil {
			t.Fatalf("todo create failed: %v, out: %s", err, out)
		}

		// 4.2 Query and extract created ID
		todosRaw, err := c.MyTodos(ctx, zentao.Params{"type": {"today"}})
		if err != nil {
			t.Fatalf("MyTodos query failed: %v", err)
		}
		var todosEnv struct {
			Todos []map[string]any `json:"todos"`
		}
		_ = json.Unmarshal(todosRaw, &todosEnv)

		var createdID string
		for _, item := range todosEnv.Todos {
			if item["name"] == uniqueName {
				createdID = fmt.Sprint(item["id"])
				break
			}
		}

		if createdID == "" {
			t.Fatalf("created todo %q not found in list", uniqueName)
		}

		// Register zero-pollution cleanup
		t.Cleanup(func() {
			_, _ = c.TodoDelete(context.Background(), createdID)
		})

		// 4.3 Start
		if _, err := execCmd("todo", "start", "--id", createdID, "-o", "json"); err != nil {
			t.Fatalf("todo start failed: %v", err)
		}

		// 4.4 Finish
		if _, err := execCmd("todo", "finish", "--id", createdID, "-o", "json"); err != nil {
			t.Fatalf("todo finish failed: %v", err)
		}

		// 4.5 Close
		if _, err := execCmd("todo", "close", "--id", createdID, "-o", "json"); err != nil {
			t.Fatalf("todo close failed: %v", err)
		}
	})

	// 5. Task Lifecycle: Params -> Create -> Finish -> Cleanup (Delete)
	t.Run("Task_Lifecycle_ZeroArtifact", func(t *testing.T) {
		targetProjectID := "109" // 8月迭代 execution ID

		// 5.1 Test Task Params
		outParams, errParams := execCmd("task", "params", "--project", targetProjectID, "-o", "json")
		if errParams != nil {
			t.Fatalf("task params failed: %v, out: %s", errParams, outParams)
		}

		// 5.2 Create Task
		taskTitle := fmt.Sprintf("【CI测试任务-%d】", time.Now().UnixNano())
		outCreate, errCreate := execCmd("task", "create", "--project", targetProjectID, "--name", taskTitle, "--assigned-to", c.Account, "--estimate", "1.0", "-o", "json")
		if errCreate != nil {
			t.Fatalf("task create failed: %v, out: %s", errCreate, outCreate)
		}

		// 5.3 Query to find created task ID (using my task openedBy)
		taskListRaw, err := c.MyTasks(ctx, zentao.Params{"type": {"openedBy"}})
		if err != nil {
			t.Fatalf("MyTasks query failed: %v", err)
		}
		var tasksEnv struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(taskListRaw, &tasksEnv)

		var createdTaskID string
		for _, tk := range tasksEnv.Tasks {
			if tk["name"] == taskTitle {
				createdTaskID = fmt.Sprint(tk["id"])
				break
			}
		}

		if createdTaskID == "" {
			t.Fatalf("created task %q not found in task list", taskTitle)
		}

		// Register cleanup
		t.Cleanup(func() {
			_, _ = c.TaskDelete(context.Background(), targetProjectID, createdTaskID)
		})

		// 5.4 Finish Task
		if _, err := execCmd("task", "finish", "--id", createdTaskID, "--real", "1.0", "--comment", "CI测试自动完成", "-o", "json"); err != nil {
			t.Fatalf("task finish failed: %v", err)
		}
	})

	// 6. Bug Lifecycle: Params -> Create -> Resolve -> Cleanup (Delete)
	t.Run("Bug_Lifecycle_ZeroArtifact", func(t *testing.T) {
		targetProductID := "8" // 天玑智脑 product ID

		// 6.1 Bug Params
		outParams, errParams := execCmd("bug", "params", "--product", targetProductID, "-o", "json")
		if errParams != nil {
			t.Fatalf("bug params failed: %v, out: %s", errParams, outParams)
		}

		// 6.2 Bug Create
		bugTitle := fmt.Sprintf("【CI测试缺陷-%d】", time.Now().UnixNano())
		outCreate, errCreate := execCmd("bug", "create", "--product", targetProductID, "--title", bugTitle, "--severity", "3", "--pri", "3", "--assigned-to", c.Account, "--steps", "1. CI测试步骤", "-o", "json")
		if errCreate != nil {
			t.Fatalf("bug create failed: %v, out: %s", errCreate, outCreate)
		}

		// 6.3 Query to find created Bug ID: in ZenTao bug/browse returns {"bugs": [...]} or map
		bugsRaw, err := c.BugList(ctx, zentao.Params{"productID": {targetProductID}, "browseType": {"all"}, "orderBy": {"id_desc"}})
		if err != nil {
			t.Fatalf("BugList query failed: %v", err)
		}

		var createdBugID string
		var genericData any
		_ = json.Unmarshal(bugsRaw, &genericData)

		if m, ok := genericData.(map[string]any); ok {
			if rawBugs, has := m["bugs"]; has {
				if bugsSlice, isSlice := rawBugs.([]any); isSlice {
					for _, item := range bugsSlice {
						if bMap, isBMap := item.(map[string]any); isBMap && bMap["title"] == bugTitle {
							createdBugID = fmt.Sprint(bMap["id"])
							break
						}
					}
				} else if bugsMap, isMap := rawBugs.(map[string]any); isMap {
					for _, item := range bugsMap {
						if bMap, isBMap := item.(map[string]any); isBMap && bMap["title"] == bugTitle {
							createdBugID = fmt.Sprint(bMap["id"])
							break
						}
					}
				}
			}
		}

		// Fallback to my bug list if needed
		if createdBugID == "" {
			myBugsRaw, _ := c.MyBugs(ctx, zentao.Params{"type": {"openedBy"}})
			var myBugsEnv struct {
				Bugs []map[string]any `json:"bugs"`
			}
			_ = json.Unmarshal(myBugsRaw, &myBugsEnv)
			for _, b := range myBugsEnv.Bugs {
				if b["title"] == bugTitle {
					createdBugID = fmt.Sprint(b["id"])
					break
				}
			}
		}

		if createdBugID == "" {
			t.Fatalf("created bug %q not found in list", bugTitle)
		}

		// Register cleanup
		t.Cleanup(func() {
			_, _ = c.BugDelete(context.Background(), createdBugID)
		})

		// 6.4 Resolve Bug
		if _, err := execCmd("bug", "resolve", "--id", createdBugID, "--resolution", "fixed", "--comment", "CI测试解决", "-o", "json"); err != nil {
			t.Fatalf("bug resolve failed: %v", err)
		}
	})

	// 7. Schema, Skill, Version Commands
	t.Run("Meta_And_System_Commands", func(t *testing.T) {
		metaCmds := [][]string{
			{"schema", "--compact", "-o", "json"},
			{"schema", "task", "--compact", "-o", "json"},
			{"schema", "bug", "--compact", "-o", "json"},
			{"schema", "my", "--compact", "-o", "json"},
			{"skill", "setup", "--target", "all", "-o", "json"},
			{"version", "-o", "json"},
			{"version", "-o", "table"},
			{"config", "show", "-o", "table"},
		}

		for _, args := range metaCmds {
			name := strings.Join(args, "_")
			t.Run(name, func(t *testing.T) {
				out, err := execCmd(args...)
				if err != nil {
					t.Fatalf("command %v failed: %v, out: %s", args, err, out)
				}
			})
		}
	})
}
