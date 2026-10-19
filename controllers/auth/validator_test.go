package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/auth"
	"github.com/personal-stakshi/draftly/backend/models"
)

func jsonContext(body string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c
}

func TestValidateAuthRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := auth.ValidateAuthRequest(jsonContext(`{"email":"a@b.com","password":"password123"}`))
		if err != nil || req.Email != "a@b.com" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		if _, err := auth.ValidateAuthRequest(jsonContext(`{"email":"not-an-email","password":"password123"}`)); err == nil {
			t.Fatal("expected an email error")
		}
	})

	t.Run("password too short", func(t *testing.T) {
		if _, err := auth.ValidateAuthRequest(jsonContext(`{"email":"a@b.com","password":"123"}`)); err == nil {
			t.Fatal("expected a min length error")
		}
	})

	t.Run("email too long", func(t *testing.T) {
		body := `{"email":"` + strings.Repeat("a", models.MaxEmailLength+1) + `@b.com","password":"password123"}`
		if _, err := auth.ValidateAuthRequest(jsonContext(body)); err == nil {
			t.Fatal("expected an email length error")
		}
	})

	t.Run("password too long", func(t *testing.T) {
		body := `{"email":"a@b.com","password":"` + strings.Repeat("a", models.MaxPasswordLength+1) + `"}`
		if _, err := auth.ValidateAuthRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a password length error")
		}
	})
}
