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
