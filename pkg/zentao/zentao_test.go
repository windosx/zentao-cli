package zentao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_Login_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		if m == "api" && f == "getSessionID" {
			resp := map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"mock123456","rand":"mockrand789"}`,
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		if m == "user" && f == "login" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			// Verify that password was NOT added to GET query string
			if r.URL.Query().Get("password") != "" {
				t.Errorf("password must not be sent in GET query string")
			}
			_ = r.ParseForm()
			if r.FormValue("account") == "testuser" && r.FormValue("password") == "Test@123456" {
				resp := map[string]any{
					"status": "success",
					"user":   map[string]any{"id": "1", "account": "testuser"},
				}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "failed", "reason": "用户名或密码错误"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := New(Config{
		URL:        server.URL,
		Account:    "testuser",
		Password:   "Test@123456",
		AccessMode: AccessModeGET,
	})

	ctx := context.Background()
	if err := client.Login(ctx); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if client.Cookie != "zentaosid=mock123456" {
		t.Errorf("unexpected cookie: %s", client.Cookie)
	}
	if client.rand != "mockrand789" {
		t.Errorf("unexpected rand: %s", client.rand)
	}
}

func TestClient_BuildURL_POST_NoBodyInQuery(t *testing.T) {
	client := New(Config{
		URL:        "http://zentao.example.com",
		AccessMode: AccessModeGET,
	})

	params := urlValues()
	params.Set("account", "testuser")
	params.Set("password", "Test@123456")

	u := client.buildURL(http.MethodPost, "user", "login", routeParam{}, params)
	if strings.Contains(u, "password") || strings.Contains(u, "testuser") {
		t.Errorf("POST request URL must not contain form body params: %s", u)
	}
	if !strings.Contains(u, "m=user") || !strings.Contains(u, "f=login") {
		t.Errorf("expected m=user and f=login in query string: %s", u)
	}
}

func TestClient_MyAndTodoAPIs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "my" && f == "task":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"1","name":"My Task"}]`})
		case m == "my" && f == "bug":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"2","title":"My Bug"}]`})
		case m == "my" && f == "todo":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"3","name":"My Todo"}]`})
		case m == "my" && f == "story":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"4","title":"My Story"}]`})
		case m == "my" && f == "project":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"5","name":"My Proj"}]`})
		case m == "my" && f == "dynamic":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": `[{"id":"6","actor":"tom"}]`})
		case m == "todo" && f == "create":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "created"})
		case m == "todo" && f == "start":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "started"})
		case m == "todo" && f == "finish":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "finished"})
		case m == "todo" && f == "close":
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "message": "closed"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(Config{URL: server.URL})
	ctx := context.Background()

	// 1. MyTasks
	if _, err := client.MyTasks(ctx, nil); err != nil {
		t.Fatalf("MyTasks failed: %v", err)
	}
	// 2. MyBugs
	if _, err := client.MyBugs(ctx, nil); err != nil {
		t.Fatalf("MyBugs failed: %v", err)
	}
	// 3. MyTodos
	if _, err := client.MyTodos(ctx, nil); err != nil {
		t.Fatalf("MyTodos failed: %v", err)
	}
	// 4. MyStories
	if _, err := client.MyStories(ctx, nil); err != nil {
		t.Fatalf("MyStories failed: %v", err)
	}
	// 5. MyProjects
	if _, err := client.MyProjects(ctx, nil); err != nil {
		t.Fatalf("MyProjects failed: %v", err)
	}
	// 6. MyDynamics
	if _, err := client.MyDynamics(ctx, nil); err != nil {
		t.Fatalf("MyDynamics failed: %v", err)
	}
	// 7. Todo CRUD
	if _, err := client.TodoCreate(ctx, Params{"name": {"Write Code"}}); err != nil {
		t.Fatalf("TodoCreate failed: %v", err)
	}
	if _, err := client.TodoStart(ctx, "3"); err != nil {
		t.Fatalf("TodoStart failed: %v", err)
	}
	if _, err := client.TodoFinish(ctx, "3"); err != nil {
		t.Fatalf("TodoFinish failed: %v", err)
	}
	if _, err := client.TodoClose(ctx, "3"); err != nil {
		t.Fatalf("TodoClose failed: %v", err)
	}
}

func TestClient_Login_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		if m == "api" && f == "getSessionID" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"sessionName":"zentaosid","sessionID":"s1","rand":"r1"}`,
			})
			return
		}
		if m == "user" && f == "login" {
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "failed", "reason": "用户名或密码错误"})
			return
		}
	}))
	defer server.Close()

	client := New(Config{
		URL:      server.URL,
		Account:  "admin",
		Password: "wrong-password",
	})

	err := client.Login(context.Background())
	if err == nil {
		t.Fatalf("expected login error, got nil")
	}
}

func TestClient_ProjectList_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		if m == "project" && f == "browse" {
			// Simulate ZenTao 12/legacy not having browse method
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body>baseRouter->triggerError('the module project has no browse method')</body></html>`))
			return
		}

		if m == "project" && f == "all" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `[{"id":"1","name":"Legacy Project"}]`,
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := New(Config{
		URL: server.URL,
	})

	data, err := client.ProjectList(context.Background(), nil)
	if err != nil {
		t.Fatalf("ProjectList fallback failed: %v", err)
	}

	var projects []map[string]any
	if err := json.Unmarshal(data, &projects); err != nil || len(projects) == 0 {
		t.Fatalf("unexpected projects data: %s", string(data))
	}
}

func TestClient_Unauthenticated_LoginPageDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return unauthenticated login template JSON
		loginView := map[string]any{
			"title":        "用户登录",
			"keepLogin":    "off",
			"loginExpired": false,
			"rand":         12345,
			"referer":      "/",
		}
		_ = json.NewEncoder(w).Encode(loginView)
	}))
	defer server.Close()

	client := New(Config{
		URL: server.URL,
	})

	_, err := client.ProjectList(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected unauthenticated error when login page is returned, got nil")
	}
}

func TestClient_TaskAndBugFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "project" && f == "task":
			resp := map[string]any{
				"status": "success",
				"data":   `[{"id":"101","name":"Design API","status":"doing"}]`,
			}
			_ = json.NewEncoder(w).Encode(resp)
		case m == "task" && f == "create":
			if r.Method == http.MethodGet {
				resp := map[string]any{"status": "success", "data": `{"projects":{"1":"Main Project"}}`}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			_ = r.ParseForm()
			if r.FormValue("name") == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resp := map[string]any{"result": "success", "message": "created"}
			_ = json.NewEncoder(w).Encode(resp)
		case m == "task" && f == "finish":
			resp := map[string]any{"result": "success", "message": "finished"}
			_ = json.NewEncoder(w).Encode(resp)
		case m == "bug" && f == "browse":
			// Verify orderBy is id_desc, not order_desc
			if r.URL.Query().Get("orderBy") == "order_desc" {
				t.Errorf("bug/browse orderBy must not be order_desc")
			}
			resp := map[string]any{
				"status": "success",
				"data":   `[{"id":"201","title":"Crash on startup","status":"active"}]`,
			}
			_ = json.NewEncoder(w).Encode(resp)
		case m == "bug" && f == "create":
			if r.Method == http.MethodGet {
				resp := map[string]any{"status": "success", "data": `{"types":{"codeerror":"Code Error"}}`}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			resp := map[string]any{"result": "success", "message": "bug created"}
			_ = json.NewEncoder(w).Encode(resp)
		case m == "bug" && f == "resolve":
			resp := map[string]any{"result": "success", "message": "bug resolved"}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(Config{
		URL: server.URL,
	})
	ctx := context.Background()

	// 1. TaskList
	tasksRaw, err := client.TaskList(ctx, Params{"project": {"1"}})
	if err != nil {
		t.Fatalf("TaskList failed: %v", err)
	}
	var tasks []map[string]any
	if err := json.Unmarshal(tasksRaw, &tasks); err != nil || len(tasks) == 0 {
		t.Fatalf("unexpected task list: %s", string(tasksRaw))
	}

	// 2. TaskCreateParams
	paramsRaw, err := client.TaskCreateParams(ctx, "1")
	if err != nil {
		t.Fatalf("TaskCreateParams failed: %v", err)
	}
	if len(paramsRaw) == 0 {
		t.Fatalf("empty task create params")
	}

	// 3. TaskCreate
	_, err = client.TaskCreate(ctx, Params{"project": {"1"}, "name": {"New Task"}})
	if err != nil {
		t.Fatalf("TaskCreate failed: %v", err)
	}

	// 4. TaskFinish
	_, err = client.TaskFinish(ctx, "101", Params{"real": {"3.5"}})
	if err != nil {
		t.Fatalf("TaskFinish failed: %v", err)
	}

	// 5. BugList
	bugsRaw, err := client.BugList(ctx, Params{"productID": {"1"}})
	if err != nil {
		t.Fatalf("BugList failed: %v", err)
	}
	var bugs []map[string]any
	if err := json.Unmarshal(bugsRaw, &bugs); err != nil || len(bugs) == 0 {
		t.Fatalf("unexpected bug list: %s", string(bugsRaw))
	}

	// 6. BugCreateParams
	bparamsRaw, err := client.BugCreateParams(ctx, "1", "0")
	if err != nil {
		t.Fatalf("BugCreateParams failed: %v", err)
	}
	if len(bparamsRaw) == 0 {
		t.Fatalf("empty bug create params")
	}

	// 7. BugCreate
	_, err = client.BugCreate(ctx, Params{"product": {"1"}, "title": {"Critical Crash"}})
	if err != nil {
		t.Fatalf("BugCreate failed: %v", err)
	}

	// 8. BugResolve
	_, err = client.BugResolve(ctx, "201", Params{"resolution": {"fixed"}})
	if err != nil {
		t.Fatalf("BugResolve failed: %v", err)
	}
}

func TestClient_OfficialSDK_AllCreateParamsAndDeletes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		switch {
		case m == "user" && f == "create" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"depts":[{"id":"1","name":"Dev"}],"rand":"userrand123"}`,
			})
		case m == "product" && f == "create" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"products":[],"poUsers":{"admin":"Admin"}}`,
			})
		case m == "project" && f == "create" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"allProducts":[{"id":"1","name":"P1"}]}`,
			})
		case m == "task" && f == "finish" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"task":{"id":"101","name":"T1"}}`,
			})
		case m == "bug" && f == "resolve" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"bug":{"id":"201","title":"B1"}}`,
			})
		case m == "task" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"result":"success"}`,
			})
		case m == "bug" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"result":"success"}`,
			})
		case m == "todo" && f == "delete":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"result":"success"}`,
			})
		case m == "dept" && f == "browse":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"tree":[{"id":"1","name":"Root Dept"}]}`,
			})
		case m == "dept" && f == "manageChild":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"data":   `{"result":"success"}`,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := New(Config{URL: server.URL})
	ctx := context.Background()

	// 1. UserCreateParams (also verifies rand update)
	uParams, err := client.UserCreateParams(ctx, "1")
	if err != nil {
		t.Fatalf("UserCreateParams failed: %v", err)
	}
	if len(uParams) == 0 || client.GetRand() != "userrand123" {
		t.Fatalf("UserCreateParams unexpected rand or empty data: %s, rand: %s", string(uParams), client.GetRand())
	}

	// 2. ProductCreateParams
	prodParams, err := client.ProductCreateParams(ctx, "0")
	if err != nil {
		t.Fatalf("ProductCreateParams failed: %v", err)
	}
	if len(prodParams) == 0 {
		t.Fatalf("ProductCreateParams empty")
	}

	// 3. ProjectCreateParams
	projParams, err := client.ProjectCreateParams(ctx, "0")
	if err != nil {
		t.Fatalf("ProjectCreateParams failed: %v", err)
	}
	if len(projParams) == 0 {
		t.Fatalf("ProjectCreateParams empty")
	}

	// 4. TaskFinishParams
	tfParams, err := client.TaskFinishParams(ctx, "101")
	if err != nil {
		t.Fatalf("TaskFinishParams failed: %v", err)
	}
	if len(tfParams) == 0 {
		t.Fatalf("TaskFinishParams empty")
	}

	// 5. BugResolveParams
	brParams, err := client.BugResolveParams(ctx, "201")
	if err != nil {
		t.Fatalf("BugResolveParams failed: %v", err)
	}
	if len(brParams) == 0 {
		t.Fatalf("BugResolveParams empty")
	}

	// 6. TaskDelete
	if _, err := client.TaskDelete(ctx, "1", "101"); err != nil {
		t.Fatalf("TaskDelete failed: %v", err)
	}

	// 7. BugDelete
	if _, err := client.BugDelete(ctx, "201"); err != nil {
		t.Fatalf("BugDelete failed: %v", err)
	}

	// 8. TodoDelete
	if _, err := client.TodoDelete(ctx, "301"); err != nil {
		t.Fatalf("TodoDelete failed: %v", err)
	}

	// 9. DeptList and DeptAdd
	if _, err := client.DeptList(ctx, nil); err != nil {
		t.Fatalf("DeptList failed: %v", err)
	}
	if _, err := client.DeptAdd(ctx, Params{"parentDeptID": {"0"}}); err != nil {
		t.Fatalf("DeptAdd failed: %v", err)
	}
}

func TestClient_ValidationErrors_And_EdgeCases(t *testing.T) {
	client := New(Config{URL: "http://localhost"})
	ctx := context.Background()

	// Task validations
	if _, err := client.TaskCreateParams(ctx, ""); err == nil {
		t.Errorf("expected error for empty projectID in TaskCreateParams")
	}
	if _, err := client.TaskCreate(ctx, Params{}); err == nil {
		t.Errorf("expected error for empty project in TaskCreate")
	}
	if _, err := client.TaskFinishParams(ctx, ""); err == nil {
		t.Errorf("expected error for empty taskID in TaskFinishParams")
	}
	if _, err := client.TaskFinish(ctx, "", nil); err == nil {
		t.Errorf("expected error for empty taskID in TaskFinish")
	}
	if _, err := client.TaskDelete(ctx, "", ""); err == nil {
		t.Errorf("expected error for empty taskID in TaskDelete")
	}

	// Bug validations
	if _, err := client.BugCreateParams(ctx, "", ""); err == nil {
		t.Errorf("expected error for empty productID in BugCreateParams")
	}
	if _, err := client.BugCreate(ctx, Params{}); err == nil {
		t.Errorf("expected error for empty product in BugCreate")
	}
	if _, err := client.BugResolveParams(ctx, ""); err == nil {
		t.Errorf("expected error for empty bugID in BugResolveParams")
	}
	if _, err := client.BugResolve(ctx, "", nil); err == nil {
		t.Errorf("expected error for empty bugID in BugResolve")
	}
	if _, err := client.BugDelete(ctx, ""); err == nil {
		t.Errorf("expected error for empty bugID in BugDelete")
	}

	// Todo validations
	if _, err := client.TodoCreate(ctx, Params{}); err == nil {
		t.Errorf("expected error for empty name in TodoCreate")
	}
	if _, err := client.TodoStart(ctx, ""); err == nil {
		t.Errorf("expected error for empty id in TodoStart")
	}
	if _, err := client.TodoFinish(ctx, ""); err == nil {
		t.Errorf("expected error for empty id in TodoFinish")
	}
	if _, err := client.TodoClose(ctx, ""); err == nil {
		t.Errorf("expected error for empty id in TodoClose")
	}
	if _, err := client.TodoDelete(ctx, ""); err == nil {
		t.Errorf("expected error for empty id in TodoDelete")
	}

	// UserAdd without rand
	if _, err := client.UserAdd(ctx, Params{}); err == nil {
		t.Errorf("expected error for UserAdd without rand")
	}
}

func TestClient_HTMLErrorAndMessageExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("m")
		f := r.URL.Query().Get("f")

		if m == "task" && f == "create" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><head><title>Error</title></head><body><h1>SQLSTATE[42000]: Syntax error or access violation</h1><p>Stack trace:</p></body></html>`))
			return
		}
		if m == "bug" && f == "create" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"fail","message":"『标题』不能为空。"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := New(Config{URL: server.URL})
	ctx := context.Background()

	// 1. HTML error
	_, err := client.TaskCreate(ctx, Params{"project": {"1"}})
	if err == nil || !strings.Contains(err.Error(), "SQLSTATE") {
		t.Fatalf("expected SQLSTATE in error message, got: %v", err)
	}

	// 2. Fail message
	_, err = client.BugCreate(ctx, Params{"product": {"1"}})
	if err == nil || !strings.Contains(err.Error(), "『标题』不能为空") {
		t.Fatalf("expected fail message in error, got: %v", err)
	}
}
