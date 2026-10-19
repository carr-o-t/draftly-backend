package share_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/share"
)

func jsonContext(body string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c
}

func TestValidateInviteRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := share.ValidateInviteRequest(jsonContext(`[{"email":"a@b.com","role":"editor"}]`))
		if err != nil || len(*req) != 1 || (*req)[0].Role != "editor" {
			t.Fatalf("expected a valid invite, got req=%v err=%v", req, err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		if _, err := share.ValidateInviteRequest(jsonContext(`[]`)); err == nil {
			t.Fatal("expected an error for an empty invite list")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		if _, err := share.ValidateInviteRequest(jsonContext(`[{"email":"nope","role":"editor"}]`)); err == nil {
			t.Fatal("expected an email error")
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		if _, err := share.ValidateInviteRequest(jsonContext(`[{"email":"a@b.com","role":"admin"}]`)); err == nil {
			t.Fatal("expected a role error")
		}
	})

	t.Run("email too long", func(t *testing.T) {
		body := `[{"email":"` + strings.Repeat("a", 300) + `@b.com","role":"editor"}]`
		if _, err := share.ValidateInviteRequest(jsonContext(body)); err == nil {
			t.Fatal("expected an email length error")
		}
	})
}

func TestValidateHandleRequest(t *testing.T) {
	t.Run("approve with role", func(t *testing.T) {
		req, err := share.ValidateHandleRequest(jsonContext(`{"action":"approved","role":"viewer"}`))
		if err != nil || req.Action != "approved" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("approve without role", func(t *testing.T) {
		if _, err := share.ValidateHandleRequest(jsonContext(`{"action":"approved"}`)); err == nil {
			t.Fatal("expected a role error on approval")
		}
	})

	t.Run("reject without role", func(t *testing.T) {
		if _, err := share.ValidateHandleRequest(jsonContext(`{"action":"rejected"}`)); err != nil {
			t.Fatalf("reject should not need a role, got %v", err)
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		if _, err := share.ValidateHandleRequest(jsonContext(`{"action":"maybe"}`)); err == nil {
			t.Fatal("expected an action error")
		}
	})
}

func TestValidateAccessTokenRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := share.ValidateAccessTokenRequest(jsonContext(`{"token":"abc"}`))
		if err != nil || req.Token != "abc" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		if _, err := share.ValidateAccessTokenRequest(jsonContext(`{}`)); err == nil {
			t.Fatal("expected a required error")
		}
	})

	t.Run("token too long", func(t *testing.T) {
		body := `{"token":"` + strings.Repeat("a", 129) + `"}`
		if _, err := share.ValidateAccessTokenRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a token length error")
		}
	})
}

func TestValidateRequestAccess(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := share.ValidateRequestAccess(jsonContext(`{"role":"viewer"}`))
		if err != nil || req.Role != "viewer" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		if _, err := share.ValidateRequestAccess(jsonContext(`{"role":"admin"}`)); err == nil {
			t.Fatal("expected a role error")
		}
	})
}
