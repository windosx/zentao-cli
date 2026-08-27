package zentao

import (
	"context"
	"crypto/md5" // #nosec G501 -- ZenTao login form protocol mandates MD5.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Common status constants.
const (
	StatusAll    = "all"
	StatusActive = "active"
	StatusClosed = "closed"
	StatusWait   = "wait"
	StatusDoing  = "doing"
	StatusDone   = "done"
	OrderIDAsc   = "id_asc"
	OrderIDDesc  = "id_desc"
	OrderAsc     = "order_asc"
	OrderDesc    = "order_desc"
)

// Params is an alias so command builders read naturally:
//
//	zentao.Params{"status": "all", "orderBy": "id_desc"}
type Params = url.Values

// ---- Department ----

// DeptList returns the department tree (m=dept&f=browse).
func (c *Client) DeptList(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"deptID": {"0"},
	}
	return c.call(ctx, http.MethodGet, "dept", "browse", mergeDefaults(params, defaults))
}

// DeptAdd creates a department (POST m=dept&f=manageChild).
func (c *Client) DeptAdd(ctx context.Context, params Params) (json.RawMessage, error) {
	return c.call(ctx, http.MethodPost, "dept", "manageChild", params)
}

// ---- User ----

// UserList returns users (m=company&f=browse).
// In ZenTao 21.7: browse(browseType='inside', param=0, type='bydept', orderBy='id_asc', recTotal=0, recPerPage=20, pageID=1)
func (c *Client) UserList(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"browseType": {"inside"},
		"param":      {"0"},
		"type":       {"bydept"},
		"orderBy":    {OrderIDAsc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	merged := mergeDefaults(params, defaults)
	return c.call(ctx, http.MethodGet, "company", "browse", merged)
}

// UserCreateParams returns parameters and schema needed to create a user (m=user&f=create&dept=<dept>).
func (c *Client) UserCreateParams(ctx context.Context, deptID string) (json.RawMessage, error) {
	params := Params{}
	if deptID != "" {
		params.Set("dept", deptID)
	}
	data, err := c.call(ctx, http.MethodGet, "user", "create", params)
	if err != nil {
		return nil, err
	}

	// Try extracting session rand if present in creation metadata
	var resp struct {
		Rand json.RawMessage `json:"rand"`
	}
	if err := json.Unmarshal(data, &resp); err == nil && len(resp.Rand) > 0 {
		r := stringValue(resp.Rand)
		if r != "" && c.rand == "" {
			c.rand = r
		}
	}

	return data, nil
}

// UserAdd creates a user (POST m=user&f=create&dept=<dept>), including the
// md5+session-rand password encryption from the PHP SDK addUser():
//
//	password1      = md5(pwd1) + rand
//	password2      = md5(pwd2) + rand
//	verifyPassword = md5(md5(masterPassword) + rand)
//
// field "password" is treated as both password1 and password2; "verifyPassword"
// defaults to the master (login) password, like the SDK.
func (c *Client) UserAdd(ctx context.Context, params Params) (json.RawMessage, error) {
	dept := params.Get("dept")
	pwd1 := params.Get("password1")
	if pwd1 == "" {
		pwd1 = params.Get("password")
	}
	pwd2 := params.Get("password2")
	if pwd2 == "" {
		pwd2 = pwd1
	}
	verify := params.Get("verifyPassword")
	if verify == "" {
		verify = c.Password
	}
	if c.rand == "" {
		return nil, fmt.Errorf("user add: session rand not available (login first)")
	}

	params = cloneParams(params)
	params.Del("password")
	params.Del("password1")
	params.Del("password2")
	params.Del("verifyPassword")
	params.Set("password1", md5hex(pwd1)+c.rand)
	params.Set("password2", md5hex(pwd2)+c.rand)
	params.Set("verifyPassword", md5hex(md5hex(verify))+c.rand)

	return c.callRoute(ctx, http.MethodPost, "user", "create", routeParam{Key: "dept", Value: dept}, params)
}

// ---- Product ----

// ProductList returns products (m=product&f=all or m=product&f=browse).
func (c *Client) ProductList(ctx context.Context, params Params) (json.RawMessage, error) {
	defaultsAll := Params{
		"browseType": {StatusAll},
		"orderBy":    {OrderDesc},
		"param":      {"0"},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	mergedAll := mergeDefaults(params, defaultsAll)
	data, err := c.call(ctx, http.MethodGet, "product", "all", mergedAll)
	if err == nil {
		return data, nil
	}

	defaultsBrowse := Params{
		"productID":  {"0"},
		"branch":     {StatusAll},
		"browseType": {StatusAll},
		"param":      {"0"},
		"storyType":  {"story"},
		"orderBy":    {OrderDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "product", "browse", mergeDefaults(params, defaultsBrowse))
}

// ProductCreateParams returns parameters and metadata needed to create a product (m=product&f=create&programID=<id>).
func (c *Client) ProductCreateParams(ctx context.Context, programID string) (json.RawMessage, error) {
	params := Params{}
	if programID != "" {
		params.Set("programID", programID)
	}
	return c.call(ctx, http.MethodGet, "product", "create", params)
}

// ProductAdd creates a product (POST m=product&f=create). Common fields:
// name, code, PO, QD, RD, acl, status, desc.
func (c *Client) ProductAdd(ctx context.Context, params Params) (json.RawMessage, error) {
	return c.call(ctx, http.MethodPost, "product", "create", params)
}

// ---- Project ----

// ProjectList returns projects.
func (c *Client) ProjectList(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"programID":  {"0"},
		"browseType": {StatusAll},
		"param":      {""},
		"orderBy":    {OrderDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	merged := mergeDefaults(params, defaults)
	if pID := params.Get("productID"); pID != "" {
		merged.Set("param", pID)
	}

	data, err := c.call(ctx, http.MethodGet, "project", "browse", merged)
	if err == nil {
		return data, nil
	}

	// Fallback to legacy project/all or execution/all
	if isModuleOrMethodError(err) {
		dataAll, errAll := c.call(ctx, http.MethodGet, "project", "all", merged)
		if errAll == nil {
			return dataAll, nil
		}
		dataExec, errExec := c.call(ctx, http.MethodGet, "execution", "all", merged)
		if errExec == nil {
			return dataExec, nil
		}
	}

	return data, err
}

// ProjectCreateParams returns parameters and metadata needed to create a project (m=project&f=create&programID=<id>).
func (c *Client) ProjectCreateParams(ctx context.Context, programID string) (json.RawMessage, error) {
	params := Params{}
	if programID != "" {
		params.Set("programID", programID)
	}
	return c.call(ctx, http.MethodGet, "project", "create", params)
}

// ProjectAdd creates a project (POST m=project&f=create). Common fields:
// name, code, begin, end, days, team, type, status, acl, PM, PO, QD, RD, desc.
func (c *Client) ProjectAdd(ctx context.Context, params Params) (json.RawMessage, error) {
	return c.call(ctx, http.MethodPost, "project", "create", params)
}

// ---- Task ----

// TaskList returns tasks of a project/execution.
func (c *Client) TaskList(ctx context.Context, params Params) (json.RawMessage, error) {
	execID := params.Get("executionID")
	if execID == "" {
		execID = params.Get("projectID")
	}
	if execID == "" {
		execID = params.Get("project")
	}

	status := params.Get("status")
	if status == "" {
		status = StatusAll
	}

	orderBy := params.Get("orderBy")
	if orderBy == "" {
		orderBy = OrderIDDesc
	}

	taskParams := Params{
		"executionID": {execID},
		"projectID":   {execID},
		"status":      {status},
		"param":       {"0"},
		"orderBy":     {orderBy},
		"recTotal":    {"999999"},
		"recPerPage":  {"999999"},
	}

	// 1. In ZenTao 21.7 tasks belong to execution/task
	data, err := c.call(ctx, http.MethodGet, "execution", "task", taskParams)
	if err == nil {
		return data, nil
	}

	// 2. Fallback to legacy project/task
	return c.call(ctx, http.MethodGet, "project", "task", taskParams)
}

// TaskCreateParams returns parameters and schema needed to create a task (m=task&f=create&project=<id>).
func (c *Client) TaskCreateParams(ctx context.Context, projectID string) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("task create params: projectID is required")
	}
	params := Params{"project": {projectID}}
	return c.call(ctx, http.MethodGet, "task", "create", params)
}

// TaskCreate creates a task (POST m=task&f=create&project=<id>). Common
// fields: name, type, pri, estimate, assignedTo, module, story, desc.
func (c *Client) TaskCreate(ctx context.Context, params Params) (json.RawMessage, error) {
	project := params.Get("project")
	if project == "" {
		project = params.Get("execution")
	}
	if project == "" {
		return nil, fmt.Errorf("task create: --project is required")
	}
	return c.callRoute(ctx, http.MethodPost, "task", "create", routeParam{Key: "project", Value: project}, params)
}

// TaskFinishParams returns parameters and current state needed to finish a task (m=task&f=finish&taskID=<id>).
func (c *Client) TaskFinishParams(ctx context.Context, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task finish params: taskID is required")
	}
	params := Params{"taskID": {taskID}}
	return c.call(ctx, http.MethodGet, "task", "finish", params)
}

// TaskFinish completes a task (POST m=task&f=finish&taskID=<id>). Common
// fields: real (actual hours), comment.
func (c *Client) TaskFinish(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task finish: --id is required")
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "finish", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskDelete deletes a task by ID (GET/POST m=task&f=delete&taskID=<id>&confirm=yes).
func (c *Client) TaskDelete(ctx context.Context, projectID, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task delete: taskID is required")
	}
	if projectID == "" {
		projectID = "0"
	}
	params := Params{
		"projectID": {projectID},
		"taskID":    {taskID},
		"confirm":   {"yes"},
	}
	return c.call(ctx, http.MethodGet, "task", "delete", params)
}

// ---- Bug ----

// BugList returns bugs (m=bug&f=browse).
func (c *Client) BugList(ctx context.Context, params Params) (json.RawMessage, error) {
	productID := params.Get("productID")
	if productID == "" {
		productID = params.Get("product")
	}
	if productID == "" {
		productID = "0"
	}

	branch := params.Get("branch")
	if branch == "" {
		branch = StatusAll
	}

	browseType := params.Get("browseType")
	if browseType == "" {
		browseType = StatusAll
	}

	orderBy := params.Get("orderBy")
	if orderBy == "" || orderBy == OrderDesc {
		orderBy = OrderIDDesc
	}

	bugParams := Params{
		"productID":  {productID},
		"branch":     {branch},
		"browseType": {browseType},
		"param":      {"0"},
		"orderBy":    {orderBy},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}

	return c.call(ctx, http.MethodGet, "bug", "browse", bugParams)
}

// BugCreateParams returns parameters and schema needed to create a bug (m=bug&f=create&productID=<id>&branch=<branch>).
func (c *Client) BugCreateParams(ctx context.Context, productID string, branch string) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("bug create params: productID is required")
	}
	if branch == "" {
		branch = "0"
	}
	params := Params{"productID": {productID}, "branch": {branch}}
	return c.call(ctx, http.MethodGet, "bug", "create", params)
}

// BugCreate creates a bug (POST m=bug&f=create&productID=<id>).
func (c *Client) BugCreate(ctx context.Context, params Params) (json.RawMessage, error) {
	product := params.Get("product")
	if product == "" {
		return nil, fmt.Errorf("bug create: --product is required")
	}

	body := cloneParams(params)
	if body.Get("status") == "" {
		body.Set("status", StatusActive)
	}
	if body.Get("type") == "" {
		body.Set("type", "codeerror")
	}
	if body.Get("severity") == "" {
		body.Set("severity", "3")
	}
	if body.Get("pri") == "" {
		body.Set("pri", "3")
	}

	// ZenTao 21.7 form::data requires openedBuild array (or openedBuild[0]=trunk)
	if len(body["openedBuild"]) == 0 && len(body["openedBuild[]"]) == 0 && len(body["openedBuild[0]"]) == 0 {
		body["openedBuild[]"] = []string{"trunk"}
		body["openedBuild[0]"] = []string{"trunk"}
		body.Set("openedBuild", "trunk")
	}

	return c.callRoute(ctx, http.MethodPost, "bug", "create", routeParam{Key: "productID", Value: product}, body)
}

// BugResolveParams returns parameters and schema needed to resolve a bug (m=bug&f=resolve&bugID=<id>).
func (c *Client) BugResolveParams(ctx context.Context, bugID string) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("bug resolve params: bugID is required")
	}
	params := Params{"bugID": {bugID}}
	return c.call(ctx, http.MethodGet, "bug", "resolve", params)
}

// BugResolve resolves a bug (POST m=bug&f=resolve&bugID=<id>). Common fields:
// resolution, resolvedBuild, comment.
func (c *Client) BugResolve(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("bug resolve: --id is required")
	}
	if params == nil {
		params = Params{}
	}
	body := cloneParams(params)
	if body.Get("status") == "" {
		body.Set("status", "resolved")
	}
	if body.Get("resolvedBuild") == "" {
		body.Set("resolvedBuild", "trunk")
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "resolve", routeParam{Key: "bugID", Value: bugID}, body)
}

// BugDelete deletes a bug by ID (GET/POST m=bug&f=delete&bugID=<id>&confirm=yes).
func (c *Client) BugDelete(ctx context.Context, bugID string) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("bug delete: bugID is required")
	}
	params := Params{
		"bugID":   {bugID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "bug", "delete", params)
}

// ---- My (Personal Workbench / Dashboard) ----

// MyTasks returns tasks for current user (m=my&f=task).
// ZenTao 21.7 signature: task(type='assignedTo', param=0, orderBy='id_desc', recTotal=0, recPerPage=20, pageID=1)
func (c *Client) MyTasks(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":       {"assignedTo"},
		"param":      {"0"},
		"orderBy":    {OrderIDDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "task", mergeDefaults(params, defaults))
}

// MyBugs returns bugs for current user (m=my&f=bug).
// ZenTao 21.7 signature: bug(type='assignedTo', param=0, orderBy='id_desc', recTotal=0, recPerPage=20, pageID=1)
func (c *Client) MyBugs(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":       {"assignedTo"},
		"param":      {"0"},
		"orderBy":    {OrderIDDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "bug", mergeDefaults(params, defaults))
}

// MyStories returns stories/requirements for current user (m=my&f=story).
// ZenTao 21.7 signature: story(type='assignedTo', param=0, orderBy='id_desc', recTotal=0, recPerPage=20, pageID=1)
func (c *Client) MyStories(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":       {"assignedTo"},
		"param":      {"0"},
		"orderBy":    {OrderIDDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "story", mergeDefaults(params, defaults))
}

// MyTodos returns todos for current user (m=my&f=todo).
// ZenTao 21.7 signature: todo(type='before', userID=”, status='all', orderBy='date_desc,status,begin', recTotal=0, recPerPage=20, pageID=1)
func (c *Client) MyTodos(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":       {StatusAll},
		"account":    {""},
		"status":     {StatusAll},
		"orderBy":    {"date_desc,status,id_desc"},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "todo", mergeDefaults(params, defaults))
}

// MyProjects returns projects for current user (m=my&f=project).
// ZenTao 21.7 signature: project(status='doing', orderBy='id_desc', recTotal=0, recPerPage=15, pageID=1)
func (c *Client) MyProjects(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"status":     {StatusAll},
		"orderBy":    {OrderDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "project", mergeDefaults(params, defaults))
}

// MyDynamics returns recent activity stream for current user (m=my&f=dynamic).
// ZenTao 21.7 signature: dynamic(type='today', recTotal=0, date=”, direction='next')
func (c *Client) MyDynamics(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":     {"today"},
		"recTotal": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "my", "dynamic", mergeDefaults(params, defaults))
}

// ---- Todo CRUD ----

// TodoList queries todo items (m=todo&f=browse or m=my&f=todo).
func (c *Client) TodoList(ctx context.Context, params Params) (json.RawMessage, error) {
	return c.MyTodos(ctx, params)
}

// TodoCreate creates a todo item (POST m=todo&f=create). Common fields:
// name, date (YYYY-MM-DD), begin, end, type (custom, task, bug, story), pri, desc.
func (c *Client) TodoCreate(ctx context.Context, params Params) (json.RawMessage, error) {
	if params.Get("name") == "" {
		return nil, fmt.Errorf("todo create: --name is required")
	}

	body := cloneParams(params)
	if body.Get("date") == "" {
		body.Set("date", time.Now().Format("2006-01-02"))
	}
	if body.Get("type") == "" {
		body.Set("type", "custom")
	}
	if body.Get("pri") == "" {
		body.Set("pri", "3")
	}
	if body.Get("status") == "" {
		body.Set("status", StatusWait)
	}
	if body.Get("begin") == "" {
		body.Set("begin", "2400")
	}
	if body.Get("end") == "" {
		body.Set("end", "2400")
	}
	if c.Account != "" {
		body.Set("account", c.Account)
	}

	dateParam := body.Get("date")
	return c.callRoute(ctx, http.MethodPost, "todo", "create", routeParam{Key: "date", Value: dateParam}, body)
}

// TodoFinish marks a todo as completed (POST m=todo&f=finish&todoID=<id>).
func (c *Client) TodoFinish(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("todo finish: --id is required")
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "finish", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoStart marks a todo as started/doing (POST m=todo&f=start&todoID=<id>).
func (c *Client) TodoStart(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("todo start: --id is required")
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "start", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoClose closes a todo (POST m=todo&f=close&todoID=<id>).
func (c *Client) TodoClose(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("todo close: --id is required")
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "close", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoDelete deletes a todo by ID (GET/POST m=todo&f=delete&todoID=<id>&confirm=yes).
func (c *Client) TodoDelete(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("todo delete: todoID is required")
	}
	params := Params{
		"todoID":  {todoID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "todo", "delete", params)
}

// ---- helpers ----

func isModuleOrMethodError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "has no") ||
		strings.Contains(s, "the module") ||
		strings.Contains(s, "not found") ||
		strings.Contains(s, "404")
}

func mergeDefaults(params Params, defaults Params) Params {
	merged := cloneParams(defaults)
	for k, vs := range params {
		merged[k] = vs
	}
	return merged
}

func cloneParams(p Params) Params {
	out := make(Params, len(p))
	for k, vs := range p {
		out[k] = append([]string(nil), vs...)
	}
	return out
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s)) // #nosec G401
	return hex.EncodeToString(sum[:])
}
