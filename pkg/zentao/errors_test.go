package zentao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorKinds(t *testing.T) {
	tests := []struct {
		name string
		err  error
		auth bool
		perm bool
	}{
		{"login page", authError("session expired or login required"), true, false},
		{"permission denied", permissionError("permission denied"), false, true},
		{"api error", apiError("operation failed"), false, false},
		{"http 401", &Error{Kind: KindAPI, HTTPStatus: 401, Message: "x"}, true, false},
		{"http 403", &Error{Kind: KindAPI, HTTPStatus: 403, Message: "x"}, false, true},
		{"wrapped auth", fmt.Errorf("wrap: %w", authError("login")), true, false},
		{"plain error", errors.New("some network issue"), false, false},
		{"nil", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.auth {
				t.Errorf("IsAuthError(%v) = %v, want %v", tt.err, got, tt.auth)
			}
			if got := IsPermissionError(tt.err); got != tt.perm {
				t.Errorf("IsPermissionError(%v) = %v, want %v", tt.err, got, tt.perm)
			}
		})
	}
}

func TestErrorMessageFormat(t *testing.T) {
	e := &Error{Kind: KindAPI, HTTPStatus: 500, Message: "boom"}
	if got := e.Error(); got != "http 500: boom" {
		t.Errorf("Error() = %q, want %q", got, "http 500: boom")
	}
	e2 := apiError("失败")
	if got := e2.Error(); got != "zentao: 失败" {
		t.Errorf("Error() = %q, want %q", got, "zentao: 失败")
	}
}

func TestValidationSentinel(t *testing.T) {
	err := fmt.Errorf("%w: task create: --project is required", ErrValidation)
	if !errors.Is(err, ErrValidation) {
		t.Error("wrapped error should match ErrValidation via errors.Is")
	}
}

func TestUnwrapResponseReturnsStructuredErrors(t *testing.T) {
	// Login page envelope -> KindAuth.
	_, err := unwrapResponse([]byte(`{"title":"用户登录","status":"success","data":"{}"}`))
	if err == nil || !IsAuthError(err) {
		t.Errorf("login page should be KindAuth, got %v", err)
	}

	// Permission denied envelope -> KindPermission.
	_, err = unwrapResponse([]byte(`{"result":"fail","load":"user-deny","message":"no priv"}`))
	if err == nil || !IsPermissionError(err) {
		t.Errorf("deny page should be KindPermission, got %v", err)
	}

	// Fail envelope -> KindAPI.
	_, err = unwrapResponse([]byte(`{"result":"fail","message":"表单校验失败"}`))
	if err == nil || IsAuthError(err) || IsPermissionError(err) {
		t.Errorf("fail envelope should be KindAPI, got %v", err)
	}
}

func TestCallRouteReLoginExcludesPermissionErrors(t *testing.T) {
	// A permission-denied response must NOT trigger transparent re-login.
	var logins int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("f") {
		case "getSessionID":
			w.Write([]byte(`{"sessionName":"zentaosid","sessionID":"s1","rand":"123"}`))
		case "login":
			logins++
			w.Write([]byte(`{"result":"success","user":{"id":1}}`))
		default:
			w.Write([]byte(`{"result":"fail","load":"user-deny","message":"no priv"}`))
		}
	}))
	defer srv.Close()

	c := New(Config{URL: srv.URL, Account: "u", Password: "p"})
	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("login: %v", err)
	}
	before := logins

	_, err := c.call(context.Background(), http.MethodGet, "task", "view", Params{})
	if err == nil || !IsPermissionError(err) {
		t.Fatalf("expected permission error, got %v", err)
	}
	if logins != before {
		t.Errorf("permission error must not trigger re-login (logins before=%d after=%d)", before, logins)
	}
}

func TestCallRouteReLoginOnAuthError(t *testing.T) {
	// First business call hits a login page; the client should re-login and retry.
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := r.URL.Query().Get("f")
		switch f {
		case "getSessionID":
			w.Write([]byte(`{"sessionName":"zentaosid","sessionID":"s1","rand":"123"}`))
		case "login":
			w.Write([]byte(`{"result":"success","user":{"id":1}}`))
		case "view":
			calls++
			if calls == 1 {
				w.Write([]byte(`{"title":"用户登录"}`))
				return
			}
			w.Write([]byte(`{"status":"success","data":"{\"ok\":true}"}`))
		default:
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	c := New(Config{URL: srv.URL, Account: "u", Password: "p"})
	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("login: %v", err)
	}
	data, err := c.call(context.Background(), http.MethodGet, "task", "view", Params{})
	if err != nil {
		t.Fatalf("call after re-login: %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("unexpected payload: %s", data)
	}
	if calls != 2 {
		t.Errorf("expected retry after re-login, calls=%d", calls)
	}
}

func TestTruncateRuneSafe(t *testing.T) {
	s := "任务创建失败的原因说明"
	got := truncate(s, 5)
	runes := []rune(got)
	// 5 runes + "..." = 8 runes, and the prefix must be valid runes.
	if string(runes[:5]) != string([]rune(s)[:5]) {
		t.Errorf("truncate broke runes: %q", got)
	}
	if got != string([]rune(s)[:5])+"..." {
		t.Errorf("truncate = %q", got)
	}
}
