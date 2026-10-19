package document_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/controllers/document"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

func newDocTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if s := c.GetHeader("X-Test-User-ID"); s != "" {
			uid, _ := uuid.Parse(s)
			c.Set("user_id", uid)
		}
		c.Next()
	})

	r.GET("/documents", document.ListDocuments)
	r.PATCH("/documents/:id/autosave", document.AutoSave)
	return r
}

// remove a test user and everything they own
func cleanupDocRows(email string) {
	var userIDs []uuid.UUID
	database.DB.Model(&models.User{}).Where("email = ?", email).Pluck("id", &userIDs)
	if len(userIDs) == 0 {
		return
	}

	var docIDs []uuid.UUID
	database.DB.Model(&models.Document{}).Where("owner_id IN ?", userIDs).Pluck("id", &docIDs)

	if len(docIDs) > 0 {
		database.DB.Where("document_id IN ?", docIDs).Delete(&models.DocumentPermission{})
		database.DB.Where("document_id IN ?", docIDs).Delete(&models.AccessRequest{})
		database.DB.Where("document_id IN ?", docIDs).Delete(&models.DocumentShare{})
		database.DB.Where("document_id IN ?", docIDs).Delete(&models.Version{})
		database.DB.Where("document_id IN ?", docIDs).Delete(&models.DocumentUpdate{})
		database.DB.Where("id IN ?", docIDs).Delete(&models.Document{})
	}

	database.DB.Where("user_id IN ?", userIDs).Delete(&models.DocumentPermission{})
	database.DB.Where("requester_id IN ?", userIDs).Delete(&models.AccessRequest{})
	database.DB.Where("email = ?", email).Delete(&models.User{})
}

func TestListDocumentsHandler(t *testing.T) {
	godotenv.Load("../../.env")
	database.ConnectDatabase(config.LoadConfig())

	ownerEmail := "test_doc_list@example.com"
	otherEmail := "test_doc_list_other@example.com"
	cleanupDocRows(ownerEmail)
	cleanupDocRows(otherEmail)
	defer cleanupDocRows(ownerEmail)
	defer cleanupDocRows(otherEmail)

	owner := models.User{Email: ownerEmail, PasswordHash: "hash", Name: "Owner"}
	other := models.User{Email: otherEmail, PasswordHash: "hash", Name: "Other"}
	database.DB.Create(&owner)
	database.DB.Create(&other)

	mineOne := models.Document{OwnerID: owner.ID, Title: "Mine One"}
	mineTwo := models.Document{OwnerID: owner.ID, Title: "Mine Two"}
	theirs := models.Document{OwnerID: other.ID, Title: "Shared With Me"}
	database.DB.Create(&mineOne)
	database.DB.Create(&mineTwo)
	database.DB.Create(&theirs)
	database.DB.Create(&models.DocumentPermission{DocumentID: theirs.ID, UserID: owner.ID, Role: models.RoleViewer})

	r := newDocTestRouter()
	req, _ := http.NewRequest("GET", "/documents", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			MyDocuments     []map[string]interface{} `json:"my_documents"`
			SharedDocuments []map[string]interface{} `json:"shared_with_me"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("bad body: %v", err)
	}
	if len(resp.Data.MyDocuments) != 2 {
		t.Fatalf("expected 2 owned docs, got %d", len(resp.Data.MyDocuments))
	}
	if len(resp.Data.SharedDocuments) != 1 {
		t.Fatalf("expected 1 shared doc, got %d", len(resp.Data.SharedDocuments))
	}
}

func TestAutoSaveHandler(t *testing.T) {
	godotenv.Load("../../.env")
	database.ConnectDatabase(config.LoadConfig())

	ownerEmail := "test_doc_autosave@example.com"
	viewerEmail := "test_doc_autosave_viewer@example.com"
	cleanupDocRows(ownerEmail)
	cleanupDocRows(viewerEmail)
	defer cleanupDocRows(ownerEmail)
	defer cleanupDocRows(viewerEmail)

	owner := models.User{Email: ownerEmail, PasswordHash: "hash", Name: "Owner"}
	viewer := models.User{Email: viewerEmail, PasswordHash: "hash", Name: "Viewer"}
	database.DB.Create(&owner)
	database.DB.Create(&viewer)

	doc := models.Document{OwnerID: owner.ID, Title: "Autosave Doc"}
	database.DB.Create(&doc)
	database.DB.Create(&models.DocumentPermission{DocumentID: doc.ID, UserID: viewer.ID, Role: models.RoleViewer})

	r := newDocTestRouter()

	t.Run("owner can autosave and content is cleaned", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"content": `<p>hi</p><script>alert(1)</script>`})
		req, _ := http.NewRequest("PATCH", "/documents/"+doc.ID.String()+"/autosave", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Data struct {
				Content string `json:"content"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if strings.Contains(resp.Data.Content, "<script") {
			t.Fatalf("response content was not sanitized: %q", resp.Data.Content)
		}

		var stored models.Document
		database.DB.First(&stored, "id = ?", doc.ID)
		if strings.Contains(stored.Content, "<script") {
			t.Fatalf("stored content was not sanitized: %q", stored.Content)
		}
	})

	t.Run("viewer cannot autosave", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"content": "<p>nope</p>"})
		req, _ := http.NewRequest("PATCH", "/documents/"+doc.ID.String()+"/autosave", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", viewer.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid document id", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"content": "x"})
		req, _ := http.NewRequest("PATCH", "/documents/not-a-uuid/autosave", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}
