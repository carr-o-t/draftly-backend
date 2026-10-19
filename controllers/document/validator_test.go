package document_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/document"
	"github.com/personal-stakshi/draftly/backend/models"
)

func jsonContext(body string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c
}

func TestValidateCreateDocumentRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := document.ValidateCreateDocumentRequest(jsonContext(`{"title":"hello"}`))
		if err != nil || req.Title != "hello" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("title too long", func(t *testing.T) {
		body := `{"title":"` + strings.Repeat("a", models.MaxTitleLength+1) + `"}`
		if _, err := document.ValidateCreateDocumentRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a too long error")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if _, err := document.ValidateCreateDocumentRequest(jsonContext(`{`)); err == nil {
			t.Fatal("expected a bind error")
		}
	})
}

func TestValidateUpdateDocumentRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := document.ValidateUpdateDocumentRequest(jsonContext(`{"title":"t","content":"<p>hi</p>"}`))
		if err != nil || req.Content != "<p>hi</p>" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("content too long", func(t *testing.T) {
		body := `{"content":"` + strings.Repeat("a", models.MaxContentLength+1) + `"}`
		if _, err := document.ValidateUpdateDocumentRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a too long error")
		}
	})

	t.Run("title too long", func(t *testing.T) {
		body := `{"title":"` + strings.Repeat("a", models.MaxTitleLength+1) + `"}`
		if _, err := document.ValidateUpdateDocumentRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a too long error")
		}
	})
}

func TestValidateAutosaveRequest(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		req, err := document.ValidateAutosaveRequest(jsonContext(`{"content":"<p>x</p>"}`))
		if err != nil || req.Content != "<p>x</p>" {
			t.Fatalf("expected a valid request, got req=%v err=%v", req, err)
		}
	})

	t.Run("missing content", func(t *testing.T) {
		if _, err := document.ValidateAutosaveRequest(jsonContext(`{}`)); err == nil {
			t.Fatal("expected a required error")
		}
	})

	t.Run("content too long", func(t *testing.T) {
		body := `{"content":"` + strings.Repeat("a", models.MaxContentLength+1) + `"}`
		if _, err := document.ValidateAutosaveRequest(jsonContext(body)); err == nil {
			t.Fatal("expected a too long error")
		}
	})
}
