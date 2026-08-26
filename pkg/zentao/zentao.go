// Package zentao implements a client for the ZenTaoPMS API, ported from the
// official PHP SDK shipped with zentaopms (zentaopms_21.7 sdk/php/zentao.php).
//
// Auth flow (mirrors the PHP SDK constructor):
//  1. GET  ?m=api&f=getSessionID            -> { sessionName, sessionID, rand }
//  2. POST ?m=user&f=login&t=json           (body: account, password)
//
// Access modes:
//   - GET (default): every request uses ?m=...&f=...&t=json query style.
//   - PATH_INFO:     POST requests go to /module-func.json instead.
//
// List (GET) responses come back as {"status":"success","data":"<json string>"}
// where data is double-encoded; submit (POST) responses come back as
// {"result":"success"|"fail","message":...}. Both are unwrapped transparently.
package zentao

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Access modes supported by the API, mirroring ztAccessMode in the PHP SDK.
const (
	AccessModeGET      = "GET"
	AccessModePathInfo = "PATH_INFO"
)

// Config holds connection settings. Resolution order: CLI flags > env vars >
// config file. See cmd config loading for details.
type Config struct {
	URL        string        `json:"url"`
	Account    string        `json:"account"`
	Password   string        `json:"password,omitempty"`
	AccessMode string        `json:"accessMode,omitempty"`
	Timeout    time.Duration `json:"-"`
	Insecure   bool          `json:"-"`
}

// Client is a ZenTao API client. Create one with New, then call Login before
// any endpoint method (Login is also invoked once per CLI invocation).
type Client struct {
	BaseURL    string
	Account    string
	Password   string
	AccessMode string

	HTTP   *http.Client
	Cookie string // sessionName=sessionID, from getSessionID
	rand   string // session random number, used by addUser password encryption

	// OnSessionRefreshed callback triggered whenever the session is automatically refreshed.
	OnSessionRefreshed func(cookie, rand string)
}

// New builds a Client from Config.
func New(cfg Config) *Client {
	transport := &http.Transport{}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit --insecure flag
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	access := strings.ToUpper(strings.TrimSpace(cfg.AccessMode))
	if access == "" {
		access = AccessModeGET
	}
	return &Client{
		BaseURL:    strings.TrimRight(cfg.URL, "/"),
		Account:    cfg.Account,
		Password:   cfg.Password,
		AccessMode: access,
		HTTP:       &http.Client{Timeout: timeout, Transport: transport},
	}
}

// GetRand returns the session rand value.
func (c *Client) GetRand() string {
	return c.rand
}

// SetRand sets the session rand value (e.g. restored from session cache).
func (c *Client) SetRand(r string) {
	c.rand = r
}

// Login performs the session handshake: getSessionID then user/login.
// It is safe to call once per process; every endpoint method assumes a valid
// session (mirrors the PHP SDK, which logs in in its constructor).
func (c *Client) Login(ctx context.Context) error {
	if c.BaseURL == "" {
		return fmt.Errorf("zentao url is required (--url / ZENTAO_URL / config)")
	}
	if c.Account == "" || c.Password == "" {
		return fmt.Errorf("zentao account and password are required (--account/--password or config file)")
	}

	// 1. session handshake.
	data, err := c.callRaw(ctx, http.MethodGet, "api", "getSessionID", routeParam{}, nil)
	if err != nil {
		return fmt.Errorf("getSessionID: %w", err)
	}
	var sid struct {
		SessionName string          `json:"sessionName"`
		SessionID   string          `json:"sessionID"`
		Rand        json.RawMessage `json:"rand"`
	}
	if err := json.Unmarshal(data, &sid); err != nil {
		return fmt.Errorf("getSessionID: parse response: %w", err)
	}
	if sid.SessionID == "" {
		return fmt.Errorf("getSessionID: empty sessionID (check url / access mode)")
	}
	c.Cookie = sid.SessionName + "=" + sid.SessionID
	c.rand = stringValue(sid.Rand)

	// 2. authenticate.
	params := urlValues()
	params.Set("account", c.Account)
	params.Set("password", c.Password)
	if _, err := c.callRaw(ctx, http.MethodPost, "user", "login", routeParam{}, params); err != nil {
		return fmt.Errorf("login: %w (account or password may be wrong)", err)
	}

	if c.OnSessionRefreshed != nil {
		c.OnSessionRefreshed(c.Cookie, c.rand)
	}
	return nil
}
