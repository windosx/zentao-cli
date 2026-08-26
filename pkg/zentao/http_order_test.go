package zentao

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClient_BuildURL_PositionalOrder(t *testing.T) {
	client := New(Config{
		URL:        "http://zentao.example.com",
		AccessMode: AccessModeGET,
	})

	// 1. Check my/todo: type -> account -> status -> orderBy -> recTotal -> recPerPage
	paramsTodo := url.Values{}
	paramsTodo.Set("type", "all")
	paramsTodo.Set("account", "")
	paramsTodo.Set("status", "all")
	paramsTodo.Set("orderBy", "date_desc,status,id_desc")
	paramsTodo.Set("recTotal", "999999")
	paramsTodo.Set("recPerPage", "999999")

	uTodo := client.buildURL(http.MethodGet, "my", "todo", routeParam{}, paramsTodo)
	expectedSubstring := "m=my&f=todo&t=json&type=all&account=&status=all&orderBy=date_desc"
	if !strings.Contains(uTodo, expectedSubstring) {
		t.Fatalf("expected positional order query substring %q in URL: %s", expectedSubstring, uTodo)
	}

	// 2. Check my/task: type -> param -> orderBy -> recTotal -> recPerPage
	paramsTask := url.Values{}
	paramsTask.Set("type", "assignedTo")
	paramsTask.Set("param", "0")
	paramsTask.Set("orderBy", "id_desc")
	paramsTask.Set("recTotal", "999999")
	paramsTask.Set("recPerPage", "999999")

	uTask := client.buildURL(http.MethodGet, "my", "task", routeParam{}, paramsTask)
	expectedTaskSub := "m=my&f=task&t=json&type=assignedTo&param=0&orderBy=id_desc&recTotal=999999&recPerPage=999999"
	if !strings.Contains(uTask, expectedTaskSub) {
		t.Fatalf("expected positional order query substring %q in URL: %s", expectedTaskSub, uTask)
	}
}

func TestClient_MyTodo_MockPositionalCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery := r.URL.RawQuery
		parts := strings.Split(rawQuery, "&")
		orderMap := make(map[string]int)
		for idx, p := range parts {
			kv := strings.SplitN(p, "=", 2)
			orderMap[kv[0]] = idx
		}

		if orderMap["type"] > orderMap["account"] ||
			orderMap["account"] > orderMap["status"] ||
			orderMap["status"] > orderMap["orderBy"] ||
			orderMap["orderBy"] > orderMap["recTotal"] {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"fail","error":"positional argument shifted"}`))
			return
		}

		_, _ = w.Write([]byte(`{"status":"success","data":"[{\"id\":\"1\",\"name\":\"Todo Item 1\",\"status\":\"wait\"}]"}`))
	}))
	defer server.Close()

	client := New(Config{URL: server.URL})
	data, err := client.MyTodos(context.Background(), nil)
	if err != nil {
		t.Fatalf("MyTodos failed: %v", err)
	}
	if !strings.Contains(string(data), "Todo Item 1") {
		t.Fatalf("unexpected MyTodos response: %s", string(data))
	}
}

func TestClient_MyTask_MockPositionalCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery := r.URL.RawQuery
		parts := strings.Split(rawQuery, "&")
		orderMap := make(map[string]int)
		for idx, p := range parts {
			kv := strings.SplitN(p, "=", 2)
			orderMap[kv[0]] = idx
		}

		// Ensure type -> param -> orderBy -> recTotal
		if orderMap["type"] > orderMap["param"] ||
			orderMap["param"] > orderMap["orderBy"] ||
			orderMap["orderBy"] > orderMap["recTotal"] {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"fail","error":"positional argument shifted"}`))
			return
		}

		_, _ = w.Write([]byte(`{"status":"success","data":"[{\"id\":\"101\",\"name\":\"My Task 101\",\"status\":\"doing\"}]"}`))
	}))
	defer server.Close()

	client := New(Config{URL: server.URL})
	data, err := client.MyTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("MyTasks failed: %v", err)
	}
	if !strings.Contains(string(data), "My Task 101") {
		t.Fatalf("unexpected MyTasks response: %s", string(data))
	}
}
