package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrinter_Success_JSON(t *testing.T) {
	var stdout bytes.Buffer
	p := New("json")
	p.Out = &stdout

	data := map[string]string{"name": "test-task", "status": "doing"}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if !resp.OK || resp.Outcome != "success" {
		t.Errorf("expected ok=true and outcome=success, got ok=%v, outcome=%s", resp.OK, resp.Outcome)
	}
}

func TestPrinter_Success_RawJSON(t *testing.T) {
	var stdout bytes.Buffer
	p := New("raw-json")
	p.Out = &stdout

	data := json.RawMessage(`{"id":123,"title":"bug"}`)
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	if !strings.Contains(stdout.String(), `"id": 123`) {
		t.Errorf("expected unmarshaled raw json, got %s", stdout.String())
	}
}

func TestPrinter_Success_YAML(t *testing.T) {
	var stdout bytes.Buffer
	p := New("yaml")
	p.Out = &stdout

	data := map[string]any{"id": 1, "name": "project"}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "ok: true") || !strings.Contains(out, "name: project") {
		t.Errorf("unexpected yaml output: %s", out)
	}
}

func TestPrinter_Success_Table_NestedCollection(t *testing.T) {
	var stdout bytes.Buffer
	p := New("table")
	p.Out = &stdout

	data := map[string]any{
		"title": "Bug List",
		"bugs": []map[string]any{
			{"id": "101", "title": "Crash on login", "status": "active", "severity": 1, "pri": 2, "assignedTo": "testuser"},
			{"id": "102", "title": "UI overflow", "status": "resolved", "severity": 3, "pri": 3, "assignedTo": "zhangsan"},
		},
	}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "TITLE") || !strings.Contains(out, "ASSIGNED_TO") {
		t.Errorf("expected table headers in output:\n%s", out)
	}
	if !strings.Contains(out, "Crash on login") || !strings.Contains(out, "testuser") {
		t.Errorf("expected rows in table output:\n%s", out)
	}
}

func TestPrinter_Success_Table_AssociativeMap(t *testing.T) {
	var stdout bytes.Buffer
	p := New("table")
	p.Out = &stdout

	data := map[string]any{
		"title": "Projects",
		"projects": map[string]any{
			"1": map[string]any{"id": "1", "name": "Sprint 1", "code": "s1", "status": "doing", "PM": "admin"},
			"2": map[string]any{"id": "2", "name": "Sprint 2", "code": "s2", "status": "wait", "PM": "admin"},
		},
	}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "PM") {
		t.Errorf("expected table headers in output:\n%s", out)
	}
	if !strings.Contains(out, "Sprint 1") || !strings.Contains(out, "Sprint 2") {
		t.Errorf("expected rows in table output:\n%s", out)
	}
}

func TestPrinter_Success_Table_KeyValue(t *testing.T) {
	var stdout bytes.Buffer
	p := New("table")
	p.Out = &stdout

	data := map[string]any{
		"account": "testuser",
		"status":  "authenticated",
		"url":     "http://zentao.example.com",
	}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "KEY") || !strings.Contains(out, "VALUE") {
		t.Errorf("expected key-value headers in output:\n%s", out)
	}
	if !strings.Contains(out, "testuser") || !strings.Contains(out, "authenticated") {
		t.Errorf("expected key-value content in output:\n%s", out)
	}
}

func TestPrinter_Success_Text_Collection(t *testing.T) {
	var stdout bytes.Buffer
	p := New("text")
	p.Out = &stdout

	data := map[string]any{
		"tasks": []map[string]any{
			{"id": "101", "name": "Implement API", "status": "doing", "pri": "2", "estimate": "4.5", "assignedTo": "testuser"},
		},
	}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "#101") || !strings.Contains(out, "Implement API") || !strings.Contains(out, "doing") {
		t.Errorf("expected text item formatting in output:\n%s", out)
	}
}

func TestPrinter_Success_Text_KeyValue(t *testing.T) {
	var stdout bytes.Buffer
	p := New("text")
	p.Out = &stdout

	data := map[string]any{
		"status":  "logged_in",
		"account": "testuser",
	}
	if err := p.Success(data); err != nil {
		t.Fatalf("Success failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "account: testuser") || !strings.Contains(out, "status: logged_in") {
		t.Errorf("expected key-value text in output:\n%s", out)
	}
}

func TestPrinter_Fail(t *testing.T) {
	var stderr bytes.Buffer
	p := New("json")
	p.Err = &stderr

	p.Fail(ExitCodeAuth, "auth", "invalid credentials", map[string]string{"account": "admin"})

	var resp Response
	if err := json.Unmarshal(stderr.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json error output: %v", err)
	}
	if resp.OK || resp.Outcome != "failure" {
		t.Errorf("expected ok=false and outcome=failure")
	}
	if resp.Error == nil || resp.Error.Code != ExitCodeAuth || resp.Error.Category != "auth" {
		t.Errorf("unexpected error info: %+v", resp.Error)
	}
}

func TestPrinter_Fail_AllFormats(t *testing.T) {
	// 1. Fail YAML
	var stderrYAML bytes.Buffer
	pYAML := New("yaml")
	pYAML.Err = &stderrYAML
	pYAML.Fail(ExitCodeValidation, "validation", "invalid param", nil)
	if !strings.Contains(stderrYAML.String(), "outcome: failure") {
		t.Errorf("expected yaml failure envelope, got %s", stderrYAML.String())
	}

	// 2. Fail Table
	var stderrTable bytes.Buffer
	pTable := New("table")
	pTable.Err = &stderrTable
	pTable.Fail(ExitCodeAPI, "api", "server down", "timeout error")
	if !strings.Contains(stderrTable.String(), "Error [api]: server down") || !strings.Contains(stderrTable.String(), "timeout error") {
		t.Errorf("expected table error output, got %s", stderrTable.String())
	}
}

func TestPrinter_Success_VariousEntities_And_Fallback(t *testing.T) {
	// 1. Todo text formatting
	pText := New("text")
	var stdoutText bytes.Buffer
	pText.Out = &stdoutText

	todoData := map[string]any{
		"todos": []map[string]any{
			{"id": "301", "name": "Write Doc", "status": "wait", "date": "2026-08-27", "begin": "0900", "end": "1800", "pri": 2, "type": "custom"},
		},
	}
	if err := pText.Success(todoData); err != nil {
		t.Fatalf("todo text failed: %v", err)
	}
	if !strings.Contains(stdoutText.String(), "#301 Write Doc") {
		t.Errorf("unexpected todo text: %s", stdoutText.String())
	}

	// 2. Story text formatting
	stdoutText.Reset()
	storyData := map[string]any{
		"stories": []map[string]any{
			{"id": "401", "title": "User login", "status": "active", "pri": "1", "estimate": "5", "assignedTo": "dev1", "openedBy": "po1"},
		},
	}
	if err := pText.Success(storyData); err != nil {
		t.Fatalf("story text failed: %v", err)
	}
	if !strings.Contains(stdoutText.String(), "#401 User login") {
		t.Errorf("unexpected story text: %s", stdoutText.String())
	}

	// 3. User and Dept table formatting
	pTable := New("table")
	var stdoutTable bytes.Buffer
	pTable.Out = &stdoutTable

	userData := map[string]any{
		"users": []map[string]any{
			{"id": "501", "account": "tom", "realname": "Tom Cat", "role": "dev", "email": "tom@test.com", "gender": "m", "mobile": "13800000000"},
		},
	}
	if err := pTable.Success(userData); err != nil {
		t.Fatalf("user table failed: %v", err)
	}
	if !strings.Contains(stdoutTable.String(), "ACCOUNT") || !strings.Contains(stdoutTable.String(), "Tom Cat") {
		t.Errorf("unexpected user table: %s", stdoutTable.String())
	}

	// 4. Fallback arbitrary columns table formatting
	stdoutTable.Reset()
	customData := []map[string]any{
		{"custom_key": "custom_val", "foo": "bar"},
	}
	if err := pTable.Success(customData); err != nil {
		t.Fatalf("custom table failed: %v", err)
	}
	if !strings.Contains(stdoutTable.String(), "CUSTOM_KEY") || !strings.Contains(stdoutTable.String(), "custom_val") {
		t.Errorf("unexpected custom table: %s", stdoutTable.String())
	}

	// 5. Empty slice & empty map
	stdoutTable.Reset()
	if err := pTable.Success([]map[string]any{}); err != nil {
		t.Fatalf("empty table failed: %v", err)
	}
	if !strings.Contains(stdoutTable.String(), "(empty list)") {
		t.Errorf("expected (empty list), got %s", stdoutTable.String())
	}

	stdoutTable.Reset()
	if err := pTable.Success(map[string]any{}); err != nil {
		t.Fatalf("empty map table failed: %v", err)
	}
	if !strings.Contains(stdoutTable.String(), "(empty)") {
		t.Errorf("expected (empty), got %s", stdoutTable.String())
	}

	// 6. Plain string in printText
	stdoutText.Reset()
	if err := pText.Success("plain string output"); err != nil {
		t.Fatalf("plain string failed: %v", err)
	}
	if !strings.Contains(stdoutText.String(), "plain string output") {
		t.Errorf("unexpected string output: %s", stdoutText.String())
	}
}
