package share_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/controllers/share"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

func setupAccessTokenRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	r.Use(func(c *gin.Context) {
		userIDStr := c.GetHeader("X-Test-User-ID")
		if userIDStr != "" {
			uid, _ := uuid.Parse(userIDStr)
			c.Set("user_id", uid)
		}
		c.Next()
	})

	r.POST("/share/access", share.AccessByToken)
	return r
}


func cleanupTestData(email string) {
	database.DB.Where("email = ?", email).Delete(&models.User{})
}

func TestAccessByToken(t *testing.T) {
	godotenv.Load("../../.env")
	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	ownerEmail := "token_owner@example.com"
	accessorEmail := "token_accessor@example.com"

	cleanupTestData(ownerEmail)
	cleanupTestData(accessorEmail)
	defer cleanupTestData(ownerEmail)
	defer cleanupTestData(accessorEmail)

	// Setup
	owner := models.User{Email: ownerEmail, Name: "Owner", PasswordHash: "hash"}
	accessor := models.User{Email: accessorEmail, Name: "Accessor", PasswordHash: "hash"}
	database.DB.Create(&owner)
	database.DB.Create(&accessor)

	doc := models.Document{OwnerID: owner.ID, Title: "Token Test Doc"}
	database.DB.Create(&doc)

	// Create share token
	share := models.DocumentShare{
		DocumentID: doc.ID,
		Token:      "test-token-12345",
		ExpiresAt:  database.DB.NowFunc().Add(24 * 365 * time.Hour),
	}
	database.DB.Create(&share)

	r := setupAccessTokenRouter()

	t.Run("Access Document by Token", func(t *testing.T) {
		reqBody := map[string]string{
			"token": "test-token-12345",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/share/access", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", accessor.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// User doesn't have permission, should get 401
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		
		if resp["success"] != false {
			t.Error("Expected success to be false")
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		reqBody := map[string]string{
			"token": "invalid-token",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/share/access", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", accessor.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			t.Error("Expected error for invalid token")
		}
	})
}
