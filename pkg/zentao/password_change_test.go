package zentao

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnwrapResponsePasswordChangeRequired(t *testing.T) {
	page := json.RawMessage(`{"title":"地盘-修改密码","isonlybody":false,"user":{"id":1,"account":"test","password":"sensitive"},"rand":123,"pager":null}`)
	encoded, err := json.Marshal(string(page))
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{
		"direct": page,
		"object data": append(append([]byte(`{"status":"success","data":`), page...), '}'),
		"string data": append(append([]byte(`{"status":"success","data":`), encoded...), '}'),
		"english": []byte(`{"title":"My-Change Password","user":{"id":1}}`),
	} {
		t.Run(name, func(t *testing.T) {
			payload, err := unwrapResponse(body)
			if !IsAuthError(err) {
				t.Fatalf("expected authentication error, got %v", err)
			}
			if len(payload) != 0 {
				t.Fatal("password change page must not be returned as business data")
			}
			if !strings.Contains(err.Error(), "password change required") {
				t.Fatalf("expected actionable password change message, got %v", err)
			}
			if strings.Contains(err.Error(), "sensitive") {
				t.Fatal("error must not expose user credentials")
			}
		})
	}
}

func TestUnwrapResponseAllowsNormalUserPayload(t *testing.T) {
	payload, err := unwrapResponse([]byte(`{"result":"success","user":{"id":1,"account":"test"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"id":1,"account":"test"}` {
		t.Fatalf("unexpected user payload: %s", payload)
	}
}
