package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/middleware"
	"github.com/personal-stakshi/draftly/backend/utils"
)

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/me", func(c *gin.Context) {
		uid := c.MustGet("user_id").(uuid.UUID)
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	return r
}

func TestAuthMiddlewareRejects(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r := setupAuthRouter()

	cases := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"no bearer prefix", "abc.def.ghi"},
		{"wrong scheme", "Token abc.def.ghi"},
		{"garbage token", "Bearer not-a-token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/me", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r := setupAuthRouter()

	userID := uuid.New()
	token, err := utils.GenerateToken(userID, &config.Config{JWTSecret: "test-secret"})
	if err != nil {
		t.Fatalf("failed to build a token: %v", err)
	}

	req, _ := http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["user_id"] != userID.String() {
		t.Fatalf("expected user %s in context, got %s", userID, resp["user_id"])
	}
}
