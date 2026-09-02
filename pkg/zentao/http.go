package zentao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// routeParam is a URL-level parameter. ZenTao's router reads route params
// (project, taskID, bugID, productID, dept) from the request URL, not from the
// POST body, so they must be placed in the query string (GET mode) or the
// path (PATH_INFO mode), exactly like the PHP SDK builds its URLs.
type routeParam struct {
	Key   string
	Value string
}

// call performs an authenticated request and returns the unwrapped payload.
func (c *Client) call(ctx context.Context, method, module, f string, params url.Values) (json.RawMessage, error) {
	return c.callRoute(ctx, method, module, f, routeParam{}, params)
}

// callRoute is call with a route parameter baked into the URL.
func (c *Client) callRoute(ctx context.Context, method, module, f string, route routeParam, params url.Values) (json.RawMessage, error) {
	data, err := c.callRaw(ctx, method, module, f, route, params)
	if err != nil && IsAuthError(err) && c.Account != "" && c.Password != "" {
		// Session expired or unauthenticated: attempt transparent re-login and retry.
		// Permission failures are excluded: re-login cannot fix them.
		if loginErr := c.Login(ctx); loginErr == nil {
			return c.callRaw(ctx, method, module, f, route, params)
		}
	}
	return data, err
}

func (c *Client) callRaw(ctx context.Context, method, module, f string, route routeParam, params url.Values) (json.RawMessage, error) {
	body, err := c.do(ctx, method, module, f, route, params)
	if err != nil {
		return nil, err
	}
	return unwrapResponse(body)
}

// do sends the HTTP request. GET requests carry params in the query string;
// POST requests carry them in the form body. The route param goes into the URL
// for POST requests (query in GET mode, path segment in PATH_INFO mode).
func (c *Client) do(ctx context.Context, method, module, f string, route routeParam, params url.Values) ([]byte, error) {
	u := c.buildURL(method, module, f, route, params)

	var req *http.Request
	var err error
	if method == http.MethodPost {
		req, err = http.NewRequestWithContext(ctx, method, u, bytes.NewReader([]byte(params.Encode())))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, u, nil)
		if err != nil {
			return nil, err
		}
	}

	// Mirror the PHP SDK headers (cookie auth + referer + X-Requested-With).
	req.Header.Set("Referer", c.BaseURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &Error{Kind: kindForStatus(resp.StatusCode), HTTPStatus: resp.StatusCode, Message: extractErrorMessage(body)}
	}
	return body, nil
}

// MethodPositionalOrder defines the exact parameter order expected by ZenTao controller action arguments.
// When constructing GET query strings, parameters matching these lists are placed FIRST in exact positional order.
// NOTE: Checked against ZenTao 21.7 source code in module/*/control.php.
var MethodPositionalOrder = map[string][]string{
	"my/todo":         {"type", "account", "status", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/task":         {"type", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/bug":          {"type", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/story":        {"type", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/project":      {"status", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/execution":    {"type", "orderBy", "recTotal", "recPerPage", "pageID"},
	"my/dynamic":      {"type", "recTotal", "date", "direction"},
	"todo/delete":     {"todoID", "confirm"},
	"task/delete":     {"projectID", "taskID", "confirm"},
	"bug/delete":      {"bugID", "confirm"},
	"story/delete":    {"storyID", "confirm"},
	"project/delete":  {"projectID", "confirm"},
	"product/delete":  {"productID", "confirm"},
	"user/delete":     {"userID", "confirm"},
	"dept/delete":     {"deptID", "confirm"},
	"action/trash":    {"type", "orderBy", "recTotal", "recPerPage", "pageID"},
	"action/undelete": {"actionID"},
	"action/hideOne":  {"actionID"},
	"action/hideAll":  {},
	"task/view":       {"taskID"},
	"bug/view":        {"bugID"},
	"story/view":      {"storyID"},
	"project/view":    {"projectID"},
	"product/view":    {"productID"},
	"todo/view":       {"todoID"},
	"user/view":       {"userID"},
	"bug/browse":      {"productID", "branch", "browseType", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"project/browse":  {"programID", "browseType", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"project/task":    {"projectID", "status", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"execution/task":  {"executionID", "status", "param", "orderBy", "recTotal", "recPerPage", "pageID"},
	"product/all":     {"browseType", "orderBy", "param", "recTotal", "recPerPage", "pageID", "programID"},
	"product/browse":  {"productID", "branch", "browseType", "param", "storyType", "orderBy", "recTotal", "recPerPage", "pageID"},
	"company/browse":  {"browseType", "param", "type", "orderBy", "recTotal", "recPerPage", "pageID"},
	"dept/browse":     {"deptID"},
}

// buildURL produces the request URL:
//
//   - GET requests:   always "<base>?m=<m>&f=<f>&t=json&<ordered_params>". ZenTao's router
//     maps parameters positionally, so positional keys are encoded in exact order.
//   - POST (GET mode):      "<base>?m=<m>&f=<f>&t=json[&<routeKey>=<route>]".
//   - POST (PATH_INFO mode): "<base>/<m>-<f>[-<route>].json".
func (c *Client) buildURL(method, module, f string, route routeParam, params url.Values) string {
	queryPairs := [][2]string{
		{"m", module},
		{"f", f},
		{"t", "json"},
	}

	if method == http.MethodPost {
		if c.AccessMode == AccessModePathInfo {
			base := strings.TrimRight(c.BaseURL, "/") + "/" + module + "-" + f
			if route.Value != "" {
				base += "-" + route.Value
			}
			return base + ".json"
		}

		// GET mode for POST requests: ONLY the route param goes in query string (not form body params)
		if route.Value != "" {
			queryPairs = append(queryPairs, [2]string{route.Key, route.Value})
		}
	} else {
		// GET requests: preserve positional argument order for ZenTao router
		key := fmt.Sprintf("%s/%s", module, f)
		order, hasOrder := MethodPositionalOrder[key]
		handled := make(map[string]bool)

		if hasOrder {
			for _, paramName := range order {
				if vs, exists := params[paramName]; exists {
					for _, v := range vs {
						queryPairs = append(queryPairs, [2]string{paramName, v})
					}
					handled[paramName] = true
				}
			}
		}

		// Append any remaining query params
		for k, vs := range params {
			if !handled[k] {
				for _, v := range vs {
					queryPairs = append(queryPairs, [2]string{k, v})
				}
			}
		}
	}

	var sb strings.Builder
	for i, pair := range queryPairs {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(url.QueryEscape(pair[0]))
		sb.WriteString("=")
		sb.WriteString(url.QueryEscape(pair[1]))
	}

	base := c.BaseURL
	if strings.Contains(base, "?") {
		return base + "&" + sb.String()
	}
	return base + "?" + sb.String()
}

type rawEnvelope struct {
	Status       string          `json:"status"`
	Result       any             `json:"result"` // can be bool false or string "fail"/"success"
	Reason       string          `json:"reason"`
	Message      json.RawMessage `json:"message"`
	Error        string          `json:"error"`
	Data         json.RawMessage `json:"data"`
	Title        string          `json:"title"`
	LoginExpired *bool           `json:"loginExpired"`
	KeepLogin    any             `json:"keepLogin"`
	Locate       string          `json:"locate"`
	Load         string          `json:"load"`
	User         json.RawMessage `json:"user"`
}

// unwrapResponse normalizes every ZenTao response shape into the payload
// json.RawMessage (the "result" field of the PHP SDK return values).
func unwrapResponse(body []byte) (json.RawMessage, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	var env rawEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		if strings.Trim(string(body), `"`) == "success" {
			return json.RawMessage(`{}`), nil
		}
		errMsg := extractErrorMessage(body)
		if isUnauthenticatedError(errMsg) {
			return nil, authError(errMsg)
		}
		return nil, apiError(errMsg)
	}

	if isLoginPage(env) {
		return nil, authError("session expired or login required (redirected to user login)")
	}

	if len(env.Data) > 0 && innerIsLoginPage(env.Data) {
		return nil, authError("session expired or login required (redirected to user login)")
	}

	if isPermissionDenied(env) {
		return nil, permissionError("permission denied (当前用户没有该操作的权限)")
	}

	if isFailureEnv(env) {
		errMsg := extractFailMessage(env)
		if isUnauthenticatedError(errMsg) {
			return nil, authError(errMsg)
		}
		return nil, apiError(errMsg)
	}

	return extractSuccessPayload(env, body)
}

func isPermissionDenied(env rawEnvelope) bool {
	return strings.Contains(env.Load, "user-deny") || strings.Contains(env.Load, "f=deny") || strings.Contains(env.Locate, "f=deny")
}

func extractSuccessPayload(env rawEnvelope, body []byte) (json.RawMessage, error) {
	if len(env.Data) > 0 {
		payload, err := decodeData(env.Data)
		if err != nil {
			return nil, fmt.Errorf("decode data: %w", err)
		}
		return payload, nil
	}

	if len(env.User) > 0 {
		return env.User, nil
	}
	if env.Status == "success" {
		return messageOrEmpty(env.Message), nil
	}
	resultStr := fmt.Sprint(env.Result)
	if resultStr == "true" || resultStr == "success" {
		return messageOrEmpty(env.Message), nil
	}
	if env.Error != "" {
		return nil, apiError(env.Error)
	}

	return body, nil
}

func isFailureEnv(env rawEnvelope) bool {
	resultStr := fmt.Sprint(env.Result)
	return env.Status == "fail" || env.Status == "failed" || env.Status == "error" ||
		resultStr == "false" || resultStr == "fail" || resultStr == "failed" || resultStr == "error"
}

func extractFailMessage(env rawEnvelope) string {
	errMsg := env.Reason
	if errMsg == "" {
		errMsg = renderMessage(env.Message)
	}
	if errMsg == "" || errMsg == "unknown error" {
		errMsg = env.Error
	}
	if errMsg == "" {
		errMsg = "operation failed"
	}
	return errMsg
}

func isLoginPage(env rawEnvelope) bool {
	if env.Title == "用户登录" || env.Title == "User Login" || env.Title == "Login" {
		return true
	}
	if env.LoginExpired != nil {
		return true
	}
	if env.KeepLogin != nil {
		return true
	}
	if env.Load == "login" || strings.Contains(env.Locate, "user-login") || strings.Contains(env.Load, "user-login") {
		return true
	}
	msg := renderMessage(env.Message)
	return isUnauthenticatedError(msg)
}

func innerIsLoginPage(data json.RawMessage) bool {
	payload, err := decodeData(data)
	if err != nil {
		return false
	}
	var inner rawEnvelope
	if err := json.Unmarshal(payload, &inner); err == nil {
		if isLoginPage(inner) {
			return true
		}
	}
	s := string(payload)
	if strings.Contains(s, `"loginExpired":true`) || strings.Contains(s, `"loginExpired":"true"`) ||
		strings.Contains(s, `"title":"用户登录"`) || strings.Contains(s, `"title":"User Login"`) {
		return true
	}
	return false
}

func isUnauthenticatedError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(msg, "登录已超时") ||
		strings.Contains(msg, "请重新登入") ||
		strings.Contains(msg, "请先登录") ||
		strings.Contains(lower, "login required") ||
		strings.Contains(lower, "session expired") ||
		strings.Contains(lower, "loginexpired") ||
		(strings.Contains(msg, "SQLSTATE[23000]") && strings.Contains(msg, "openedBy"))
}

// decodeData unwraps the double-encoded "data" field: when raw is a quoted
// JSON string it decodes it once and returns the inner JSON; otherwise the
// value is returned unchanged.
func decodeData(raw json.RawMessage) (json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		return json.RawMessage(s), nil
	}
	return raw, nil
}

// messageOrEmpty returns message when present (e.g. "success" or a string like
// "创建成功") and an empty object otherwise.
func messageOrEmpty(m json.RawMessage) json.RawMessage {
	m = bytes.TrimSpace(m)
	if len(m) == 0 {
		return json.RawMessage(`{}`)
	}
	return m
}

// renderMessage renders a "message" field for error output: unquotes plain
// strings, keeps JSON objects compact (ZenTao form validation errors).
func renderMessage(m json.RawMessage) string {
	m = bytes.TrimSpace(m)
	if len(m) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(m, &s) == nil && s != "" {
		return s
	}
	return string(m)
}

// stringValue extracts a string from a json.RawMessage that may hold either a
// JSON string or a JSON number (e.g. "rand").
func stringValue(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	if raw[0] == '"' {
		return strings.Trim(string(raw), `"`)
	}
	return string(raw)
}

func urlValues() url.Values {
	return url.Values{}
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	jsAlertRegex = regexp.MustCompile(`(?s)(?:window\.)?alert\s*\(\s*(?:'([^']*)'|"([^"]*)")\s*\)`)
)

func stripHTML(src string) string {
	return htmlTagRegex.ReplaceAllString(src, "")
}

func extractErrorMessage(body []byte) string {
	s := string(body)

	// Extract triggerError('...') content if present
	if idx := strings.Index(s, "triggerError('"); idx != -1 {
		rest := s[idx+len("triggerError('"):]
		if endIdx := strings.Index(rest, "'"); endIdx != -1 {
			return rest[:endIdx]
		}
	}

	// Extract alert('...') / window.alert('...') content if present
	if match := jsAlertRegex.FindStringSubmatch(s); len(match) > 2 {
		msg := match[1]
		if msg == "" && len(match) > 2 {
			msg = match[2]
		}
		if strings.TrimSpace(msg) != "" {
			msg = strings.TrimSpace(msg)
			msg = strings.ReplaceAll(msg, `\n`, "\n")
			msg = strings.ReplaceAll(msg, `\'`, "'")
			msg = strings.ReplaceAll(msg, `\"`, "\"")
			return cleanExtractedMessage(msg)
		}
	}

	// Extract SQLSTATE message if present
	if idx := strings.Index(s, "SQLSTATE["); idx != -1 {
		rest := s[idx:]
		if endIdx := strings.Index(rest, "<"); endIdx != -1 {
			return rest[:endIdx]
		}
		if endIdx := strings.Index(rest, "\n"); endIdx != -1 {
			return rest[:endIdx]
		}
		return truncate(rest, 200)
	}

	// Strip HTML tags and clean whitespace
	clean := stripHTML(s)
	clean = strings.Join(strings.Fields(clean), " ")
	if clean != "" {
		return truncate(clean, 300)
	}
	return truncate(s, 200)
}

func cleanExtractedMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	var cleaned []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			cleaned = append(cleaned, t)
		}
	}
	if len(cleaned) == 0 {
		return msg
	}
	return strings.Join(cleaned, " ")
}

// kindForStatus maps an HTTP status code to an ErrorKind.
func kindForStatus(status int) ErrorKind {
	switch status {
	case http.StatusUnauthorized:
		return KindAuth
	case http.StatusForbidden:
		return KindPermission
	default:
		return KindAPI
	}
}
