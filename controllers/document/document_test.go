package document_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/controllers/document"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

// Setup Router
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	// Mock Auth Middleware by setting user_id in context
	r.Use(func(c *gin.Context) {
		// Default to Test User ID
		// In real usage we might want to vary this, but for simple CRUD tests we can assume "current user" is the test user
		// We can inject it via header if needed, but for simplicity:
		userIDStr := c.GetHeader("X-Test-User-ID")
		if userIDStr != "" {
			uid, _ := uuid.Parse(userIDStr)
			c.Set("user_id", uid)
		}
		c.Next()
	})

	r.POST("/documents", document.CreateDocument)
	r.GET("/documents/:id", document.GetDocument)
	r.PATCH("/documents/:id", document.UpdateDocument)
	r.DELETE("/documents/:id", document.DeleteDocument)
	return r
}

func cleanupTestData(email string) {
	// With OnDelete:CASCADE, deleting user should delete documents automatically
	database.DB.Where("email = ?", email).Delete(&models.User{})
}

func TestDocumentCRUD(t *testing.T) {
	godotenv.Load("../../.env")
	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	testEmail := "test_doc_owner@example.com"
	cleanupTestData(testEmail)
	defer cleanupTestData(testEmail)

	// Create a dummy user for testing
	user := models.User{
		Email:        testEmail,
		PasswordHash: "hash",
		Name:         "Doc Owner",
	}
	database.DB.Create(&user)
	
	r := setupRouter()
	
	var createdDocID string

	// 1. Create Document
	t.Run("Create Document", func(t *testing.T) {
		reqBody := map[string]string{
			"title": "My Test Doc",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/documents", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", user.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		createdDocID = data["id"].(string)
		
		if data["title"] != "My Test Doc" {
			t.Error("Title mismatch")
		}
	})

	// 2. Get Document
	t.Run("Get Document", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/documents/"+createdDocID, nil)
		req.Header.Set("X-Test-User-ID", user.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	// 3. Update Document
	t.Run("Update Document", func(t *testing.T) {
		reqBody := map[string]string{
			"title":   "Updated Title",
			"content": "New Content",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", "/documents/"+createdDocID, bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", user.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		
		// Verify update
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		if data["title"] != "Updated Title" {
			t.Error("Title not updated")
		}
		if data["content"] != "New Content" {
			t.Error("Content not updated")
		}
	})

	// 4. Delete Document
	t.Run("Delete Document", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/documents/"+createdDocID, nil)
		req.Header.Set("X-Test-User-ID", user.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		
		// Verify deletion
		req2, _ := http.NewRequest("GET", "/documents/"+createdDocID, nil)
		req2.Header.Set("X-Test-User-ID", user.ID.String())
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("Expected 404 after delete, got %d", w2.Code)
		}
	})
}
