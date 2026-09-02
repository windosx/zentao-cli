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

// DeptEdit edits a department (POST m=dept&f=edit&deptID=<id>).
func (c *Client) DeptEdit(ctx context.Context, deptID string, params Params) (json.RawMessage, error) {
	if deptID == "" {
		return nil, fmt.Errorf("%w: dept edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "dept", "edit", routeParam{Key: "deptID", Value: deptID}, params)
}

// DeptDelete deletes a department (GET/POST m=dept&f=delete&deptID=<id>&confirm=yes).
func (c *Client) DeptDelete(ctx context.Context, deptID string) (json.RawMessage, error) {
	if deptID == "" {
		return nil, fmt.Errorf("%w: dept delete: deptID is required", ErrValidation)
	}
	params := Params{
		"deptID":  {deptID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "dept", "delete", params)
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

// UserView returns details of a user (GET m=user&f=view&userID=<id>).
func (c *Client) UserView(ctx context.Context, userID string) (json.RawMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user view: userID is required", ErrValidation)
	}
	params := Params{"userID": {userID}}
	return c.call(ctx, http.MethodGet, "user", "view", params)
}

// UserEditParams returns parameters and metadata needed to edit a user (m=user&f=edit&userID=<id>).
func (c *Client) UserEditParams(ctx context.Context, userID string) (json.RawMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user edit params: userID is required", ErrValidation)
	}
	params := Params{"userID": {userID}}
	return c.call(ctx, http.MethodGet, "user", "edit", params)
}

// UserEdit updates a user (POST m=user&f=edit&userID=<id>).
func (c *Client) UserEdit(ctx context.Context, userID string, params Params) (json.RawMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "user", "edit", routeParam{Key: "userID", Value: userID}, params)
}

// UserDelete deletes a user (GET/POST m=user&f=delete&userID=<id>&confirm=yes).
func (c *Client) UserDelete(ctx context.Context, userID string) (json.RawMessage, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user delete: userID is required", ErrValidation)
	}
	params := Params{
		"userID":  {userID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "user", "delete", params)
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
		return nil, authError("user add: session rand not available (login first)")
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

	// Only fall back to the legacy product/browse route when product/all does
	// not exist on this server; other errors (auth, permission, server errors)
	// are returned as-is.
	if !isModuleOrMethodError(err) {
		return nil, err
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

// ProductView returns details of a product (GET m=product&f=view&productID=<id>).
func (c *Client) ProductView(ctx context.Context, productID string) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product view: productID is required", ErrValidation)
	}
	params := Params{"productID": {productID}}
	return c.call(ctx, http.MethodGet, "product", "view", params)
}

// ProductEditParams returns parameters and metadata needed to edit a product (m=product&f=edit&productID=<id>).
func (c *Client) ProductEditParams(ctx context.Context, productID string) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product edit params: productID is required", ErrValidation)
	}
	params := Params{"productID": {productID}}
	return c.call(ctx, http.MethodGet, "product", "edit", params)
}

// ProductEdit updates a product (POST m=product&f=edit&productID=<id>).
func (c *Client) ProductEdit(ctx context.Context, productID string, params Params) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "product", "edit", routeParam{Key: "productID", Value: productID}, params)
}

// ProductClose closes a product (POST m=product&f=close&productID=<id>).
func (c *Client) ProductClose(ctx context.Context, productID string, params Params) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product close: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "product", "close", routeParam{Key: "productID", Value: productID}, params)
}

// ProductActivate activates a closed product (POST m=product&f=activate&productID=<id>).
func (c *Client) ProductActivate(ctx context.Context, productID string, params Params) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "product", "activate", routeParam{Key: "productID", Value: productID}, params)
}

// ProductDelete deletes a product (GET/POST m=product&f=delete&productID=<id>&confirm=yes).
func (c *Client) ProductDelete(ctx context.Context, productID string) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: product delete: productID is required", ErrValidation)
	}
	params := Params{
		"productID": {productID},
		"confirm":   {"yes"},
	}
	return c.call(ctx, http.MethodGet, "product", "delete", params)
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

// ProjectView returns details of a project (GET m=project&f=view&projectID=<id>).
func (c *Client) ProjectView(ctx context.Context, projectID string) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project view: projectID is required", ErrValidation)
	}
	params := Params{"projectID": {projectID}}
	return c.call(ctx, http.MethodGet, "project", "view", params)
}

// ProjectEditParams returns parameters and metadata needed to edit a project (m=project&f=edit&projectID=<id>).
func (c *Client) ProjectEditParams(ctx context.Context, projectID string) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project edit params: projectID is required", ErrValidation)
	}
	params := Params{"projectID": {projectID}}
	return c.call(ctx, http.MethodGet, "project", "edit", params)
}

// ProjectEdit updates a project (POST m=project&f=edit&projectID=<id>).
func (c *Client) ProjectEdit(ctx context.Context, projectID string, params Params) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "project", "edit", routeParam{Key: "projectID", Value: projectID}, params)
}

// ProjectStart starts a project (POST m=project&f=start&projectID=<id>).
func (c *Client) ProjectStart(ctx context.Context, projectID string, params Params) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project start: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "project", "start", routeParam{Key: "projectID", Value: projectID}, params)
}

// ProjectSuspend suspends a project (POST m=project&f=suspend&projectID=<id>).
func (c *Client) ProjectSuspend(ctx context.Context, projectID string, params Params) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project suspend: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "project", "suspend", routeParam{Key: "projectID", Value: projectID}, params)
}

// ProjectActivate activates a suspended or closed project (POST m=project&f=activate&projectID=<id>).
func (c *Client) ProjectActivate(ctx context.Context, projectID string, params Params) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "project", "activate", routeParam{Key: "projectID", Value: projectID}, params)
}

// ProjectClose closes a project (POST m=project&f=close&projectID=<id>).
func (c *Client) ProjectClose(ctx context.Context, projectID string, params Params) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project close: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "project", "close", routeParam{Key: "projectID", Value: projectID}, params)
}

// ProjectDelete deletes a project (GET/POST m=project&f=delete&projectID=<id>&confirm=yes).
func (c *Client) ProjectDelete(ctx context.Context, projectID string) (json.RawMessage, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project delete: projectID is required", ErrValidation)
	}
	params := Params{
		"projectID": {projectID},
		"confirm":   {"yes"},
	}
	return c.call(ctx, http.MethodGet, "project", "delete", params)
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
		return nil, fmt.Errorf("%w: task create params: projectID is required", ErrValidation)
	}
	params := Params{"project": {projectID}}
	return c.call(ctx, http.MethodGet, "task", "create", params)
}

// TaskCreate creates a task (POST m=task&f=create&executionID=<id>). Common
// fields: name, type, pri, estimate, assignedTo, module, story, desc, keywords, mailto, deadline, estStarted.
func (c *Client) TaskCreate(ctx context.Context, params Params) (json.RawMessage, error) {
	project := params.Get("executionID")
	if project == "" {
		project = params.Get("execution")
	}
	if project == "" {
		project = params.Get("project")
	}
	if project == "" {
		return nil, fmt.Errorf("%w: task create: --project is required", ErrValidation)
	}
	// Pass execution in body as well for ZenTao 21.7 form processor
	if params.Get("execution") == "" {
		params.Set("execution", project)
	}
	return c.callRoute(ctx, http.MethodPost, "task", "create", routeParam{Key: "executionID", Value: project}, params)
}

// TaskView returns details of a task (GET m=task&f=view&taskID=<id>).
func (c *Client) TaskView(ctx context.Context, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task view: taskID is required", ErrValidation)
	}
	params := Params{"taskID": {taskID}}
	return c.call(ctx, http.MethodGet, "task", "view", params)
}

// TaskEditParams returns parameters and schema needed to edit a task (m=task&f=edit&taskID=<id>).
func (c *Client) TaskEditParams(ctx context.Context, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task edit params: taskID is required", ErrValidation)
	}
	params := Params{"taskID": {taskID}}
	return c.call(ctx, http.MethodGet, "task", "edit", params)
}

// TaskEdit updates a task (POST m=task&f=edit&taskID=<id>).
func (c *Client) TaskEdit(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "edit", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskStart starts a task (POST m=task&f=start&taskID=<id>).
func (c *Client) TaskStart(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task start: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "start", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskPause pauses a task (POST m=task&f=pause&taskID=<id>).
func (c *Client) TaskPause(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task pause: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "pause", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskRestart restarts a paused task (POST m=task&f=restart&taskID=<id>).
func (c *Client) TaskRestart(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task restart: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "restart", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskClose closes a completed or canceled task (POST m=task&f=close&taskID=<id>).
func (c *Client) TaskClose(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task close: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "close", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskCancel cancels a task (POST m=task&f=cancel&taskID=<id>).
func (c *Client) TaskCancel(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task cancel: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "cancel", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskActivate activates a task (POST m=task&f=activate&taskID=<id>).
func (c *Client) TaskActivate(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "activate", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskAssign assigns a task to a user (POST m=task&f=assignTo&taskID=<id>).
func (c *Client) TaskAssign(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task assign: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "assignTo", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskFinishParams returns parameters and current state needed to finish a task (m=task&f=finish&taskID=<id>).
func (c *Client) TaskFinishParams(ctx context.Context, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task finish params: taskID is required", ErrValidation)
	}
	params := Params{"taskID": {taskID}}
	return c.call(ctx, http.MethodGet, "task", "finish", params)
}

// TaskFinish completes a task (POST m=task&f=finish&taskID=<id>). Common
// fields: real (actual hours), comment.
func (c *Client) TaskFinish(ctx context.Context, taskID string, params Params) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task finish: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "task", "finish", routeParam{Key: "taskID", Value: taskID}, params)
}

// TaskDelete deletes a task by ID (GET/POST m=task&f=delete&taskID=<id>&confirm=yes).
func (c *Client) TaskDelete(ctx context.Context, projectID, taskID string) (json.RawMessage, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task delete: taskID is required", ErrValidation)
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
		return nil, fmt.Errorf("%w: bug create params: productID is required", ErrValidation)
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
		return nil, fmt.Errorf("%w: bug create: --product is required", ErrValidation)
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

// BugView returns details of a bug (GET m=bug&f=view&bugID=<id>).
func (c *Client) BugView(ctx context.Context, bugID string) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug view: bugID is required", ErrValidation)
	}
	params := Params{"bugID": {bugID}}
	return c.call(ctx, http.MethodGet, "bug", "view", params)
}

// BugEditParams returns parameters and schema needed to edit a bug (m=bug&f=edit&bugID=<id>).
func (c *Client) BugEditParams(ctx context.Context, bugID string) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug edit params: bugID is required", ErrValidation)
	}
	params := Params{"bugID": {bugID}}
	return c.call(ctx, http.MethodGet, "bug", "edit", params)
}

// BugEdit updates a bug (POST m=bug&f=edit&bugID=<id>).
func (c *Client) BugEdit(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "edit", routeParam{Key: "bugID", Value: bugID}, params)
}

// BugClose closes a resolved or active bug (POST m=bug&f=close&bugID=<id>).
func (c *Client) BugClose(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug close: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "close", routeParam{Key: "bugID", Value: bugID}, params)
}

// BugActivate activates a resolved or closed bug (POST m=bug&f=activate&bugID=<id>).
func (c *Client) BugActivate(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "activate", routeParam{Key: "bugID", Value: bugID}, params)
}

// BugAssign assigns a bug to a user (POST m=bug&f=assignTo&bugID=<id>).
func (c *Client) BugAssign(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug assign: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "assignTo", routeParam{Key: "bugID", Value: bugID}, params)
}

// BugConfirm confirms a bug (POST m=bug&f=confirmBug&bugID=<id>).
func (c *Client) BugConfirm(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug confirm: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "bug", "confirmBug", routeParam{Key: "bugID", Value: bugID}, params)
}

// BugResolveParams returns parameters and schema needed to resolve a bug (m=bug&f=resolve&bugID=<id>).
func (c *Client) BugResolveParams(ctx context.Context, bugID string) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug resolve params: bugID is required", ErrValidation)
	}
	params := Params{"bugID": {bugID}}
	return c.call(ctx, http.MethodGet, "bug", "resolve", params)
}

// BugResolve resolves a bug (POST m=bug&f=resolve&bugID=<id>). Common fields:
// resolution, resolvedBuild, comment.
func (c *Client) BugResolve(ctx context.Context, bugID string, params Params) (json.RawMessage, error) {
	if bugID == "" {
		return nil, fmt.Errorf("%w: bug resolve: --id is required", ErrValidation)
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
		return nil, fmt.Errorf("%w: bug delete: bugID is required", ErrValidation)
	}
	params := Params{
		"bugID":   {bugID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "bug", "delete", params)
}

// ---- Story / Requirement ----

// StoryList returns stories for a product (m=product&f=browse&storyType=story).
func (c *Client) StoryList(ctx context.Context, params Params) (json.RawMessage, error) {
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
		browseType = "unclosed"
	}

	storyType := params.Get("storyType")
	if storyType == "" {
		storyType = "story"
	}

	orderBy := params.Get("orderBy")
	if orderBy == "" || orderBy == OrderDesc {
		orderBy = OrderIDDesc
	}

	storyParams := Params{
		"productID":  {productID},
		"branch":     {branch},
		"browseType": {browseType},
		"param":      {"0"},
		"storyType":  {storyType},
		"orderBy":    {orderBy},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}

	return c.call(ctx, http.MethodGet, "product", "browse", storyParams)
}

// StoryView returns details of a story (GET m=story&f=view&storyID=<id>).
func (c *Client) StoryView(ctx context.Context, storyID string) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story view: storyID is required", ErrValidation)
	}
	params := Params{"storyID": {storyID}}
	return c.call(ctx, http.MethodGet, "story", "view", params)
}

// StoryCreateParams returns parameters and schema needed to create a story (m=story&f=create&productID=<id>&branch=<b>).
func (c *Client) StoryCreateParams(ctx context.Context, productID, branch string) (json.RawMessage, error) {
	if productID == "" {
		return nil, fmt.Errorf("%w: story create params: productID is required", ErrValidation)
	}
	if branch == "" {
		branch = "0"
	}
	params := Params{"productID": {productID}, "branch": {branch}}
	return c.call(ctx, http.MethodGet, "story", "create", params)
}

// StoryCreate creates a story (POST m=story&f=create&productID=<id>&branch=<b>).
func (c *Client) StoryCreate(ctx context.Context, params Params) (json.RawMessage, error) {
	product := params.Get("product")
	if product == "" {
		product = params.Get("productID")
	}
	if product == "" {
		return nil, fmt.Errorf("%w: story create: --product is required", ErrValidation)
	}

	body := cloneParams(params)
	if body.Get("type") == "" {
		body.Set("type", "story")
	}
	if body.Get("pri") == "" {
		body.Set("pri", "3")
	}

	return c.callRoute(ctx, http.MethodPost, "story", "create", routeParam{Key: "productID", Value: product}, body)
}

// StoryEditParams returns parameters and schema needed to edit a story (m=story&f=edit&storyID=<id>).
func (c *Client) StoryEditParams(ctx context.Context, storyID string) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story edit params: storyID is required", ErrValidation)
	}
	params := Params{"storyID": {storyID}}
	return c.call(ctx, http.MethodGet, "story", "edit", params)
}

// StoryEdit updates a story (POST m=story&f=edit&storyID=<id>).
func (c *Client) StoryEdit(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "edit", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryReview reviews a story (POST m=story&f=review&storyID=<id>).
func (c *Client) StoryReview(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story review: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "review", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryChange changes a story's specification/details (POST m=story&f=change&storyID=<id>).
func (c *Client) StoryChange(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story change: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "change", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryClose closes a story (POST m=story&f=close&storyID=<id>).
func (c *Client) StoryClose(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story close: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "close", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryActivate activates a story (POST m=story&f=activate&storyID=<id>).
func (c *Client) StoryActivate(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "activate", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryAssign assigns a story to a user (POST m=story&f=assignTo&storyID=<id>).
func (c *Client) StoryAssign(ctx context.Context, storyID string, params Params) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story assign: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "story", "assignTo", routeParam{Key: "storyID", Value: storyID}, params)
}

// StoryDelete deletes a story (GET/POST m=story&f=delete&storyID=<id>&confirm=yes).
func (c *Client) StoryDelete(ctx context.Context, storyID string) (json.RawMessage, error) {
	if storyID == "" {
		return nil, fmt.Errorf("%w: story delete: storyID is required", ErrValidation)
	}
	params := Params{
		"storyID": {storyID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "story", "delete", params)
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
		return nil, fmt.Errorf("%w: todo create: --name is required", ErrValidation)
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
		return nil, fmt.Errorf("%w: todo finish: --id is required", ErrValidation)
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "finish", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoStart marks a todo as started/doing (POST m=todo&f=start&todoID=<id>).
func (c *Client) TodoStart(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo start: --id is required", ErrValidation)
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "start", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoClose closes a todo (POST m=todo&f=close&todoID=<id>).
func (c *Client) TodoClose(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo close: --id is required", ErrValidation)
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "close", routeParam{Key: "todoID", Value: todoID}, Params{})
}

// TodoView returns details of a todo item (GET m=todo&f=view&todoID=<id>).
func (c *Client) TodoView(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo view: todoID is required", ErrValidation)
	}
	params := Params{"todoID": {todoID}}
	return c.call(ctx, http.MethodGet, "todo", "view", params)
}

// TodoEditParams returns parameters and metadata needed to edit a todo (m=todo&f=edit&todoID=<id>).
func (c *Client) TodoEditParams(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo edit params: todoID is required", ErrValidation)
	}
	params := Params{"todoID": {todoID}}
	return c.call(ctx, http.MethodGet, "todo", "edit", params)
}

// TodoEdit updates a todo item (POST m=todo&f=edit&todoID=<id>).
func (c *Client) TodoEdit(ctx context.Context, todoID string, params Params) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo edit: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "edit", routeParam{Key: "todoID", Value: todoID}, params)
}

// TodoActivate activates a completed/closed todo (POST m=todo&f=activate&todoID=<id>).
func (c *Client) TodoActivate(ctx context.Context, todoID string, params Params) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo activate: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "activate", routeParam{Key: "todoID", Value: todoID}, params)
}

// TodoAssign assigns a todo to another user (POST m=todo&f=assignTo&todoID=<id>).
func (c *Client) TodoAssign(ctx context.Context, todoID string, params Params) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo assign: --id is required", ErrValidation)
	}
	if params == nil {
		params = Params{}
	}
	return c.callRoute(ctx, http.MethodPost, "todo", "assignTo", routeParam{Key: "todoID", Value: todoID}, params)
}

// TodoDelete deletes a todo by ID (GET/POST m=todo&f=delete&todoID=<id>&confirm=yes).
func (c *Client) TodoDelete(ctx context.Context, todoID string) (json.RawMessage, error) {
	if todoID == "" {
		return nil, fmt.Errorf("%w: todo delete: todoID is required", ErrValidation)
	}
	params := Params{
		"todoID":  {todoID},
		"confirm": {"yes"},
	}
	return c.call(ctx, http.MethodGet, "todo", "delete", params)
}

// ---- Trash & Recycle Bin (Action Module) ----

// TrashList queries items in the recycle bin (m=action&f=trash).
func (c *Client) TrashList(ctx context.Context, params Params) (json.RawMessage, error) {
	defaults := Params{
		"type":       {"all"},
		"orderBy":    {OrderIDDesc},
		"recTotal":   {"999999"},
		"recPerPage": {"999999"},
	}
	return c.call(ctx, http.MethodGet, "action", "trash", mergeDefaults(params, defaults))
}

// TrashRestore restores a deleted object from the recycle bin by its action ID (m=action&f=undelete&actionID=<id>).
func (c *Client) TrashRestore(ctx context.Context, actionID string) (json.RawMessage, error) {
	if actionID == "" {
		return nil, fmt.Errorf("%w: trash restore: actionID is required", ErrValidation)
	}
	params := Params{"actionID": {actionID}}
	return c.call(ctx, http.MethodGet, "action", "undelete", params)
}

// TrashHideOne hides an item in the recycle bin by action ID (m=action&f=hideOne&actionID=<id>).
func (c *Client) TrashHideOne(ctx context.Context, actionID string) (json.RawMessage, error) {
	if actionID == "" {
		return nil, fmt.Errorf("%w: trash hide-one: actionID is required", ErrValidation)
	}
	params := Params{"actionID": {actionID}}
	return c.call(ctx, http.MethodGet, "action", "hideOne", params)
}

// TrashHideAll hides all items in the recycle bin (m=action&f=hideAll).
func (c *Client) TrashHideAll(ctx context.Context) (json.RawMessage, error) {
	return c.call(ctx, http.MethodGet, "action", "hideAll", Params{})
}

// RestoreObject restores a deleted object of a given type and ID by locating its deletion action in the trash.
func (c *Client) RestoreObject(ctx context.Context, objectType, objectID string) (json.RawMessage, error) {
	if objectType == "" || objectID == "" {
		return nil, fmt.Errorf("%w: restore object: objectType and objectID are required", ErrValidation)
	}

	trashData, err := c.TrashList(ctx, Params{"type": {objectType}})
	if err != nil {
		return nil, fmt.Errorf("fetch trash list: %w", err)
	}

	var resp struct {
		Trashes []struct {
			ID         any    `json:"id"`
			ObjectType string `json:"objectType"`
			ObjectID   any    `json:"objectID"`
		} `json:"trashes"`
	}
	if err := json.Unmarshal(trashData, &resp); err != nil {
		return nil, fmt.Errorf("parse trash list: %w", err)
	}

	var matchedActionID string
	for _, t := range resp.Trashes {
		tObjID := fmt.Sprint(t.ObjectID)
		if strings.EqualFold(t.ObjectType, objectType) && tObjID == objectID {
			matchedActionID = fmt.Sprint(t.ID)
			break
		}
	}

	if matchedActionID == "" {
		return nil, fmt.Errorf("object %s #%s not found in trash (it may already be restored or permanently purged)", objectType, objectID)
	}

	return c.TrashRestore(ctx, matchedActionID)
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
