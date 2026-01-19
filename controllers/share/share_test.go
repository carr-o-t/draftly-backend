package share_test

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
	"github.com/personal-stakshi/draftly/backend/controllers/share"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

// Setup Router
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	// Mock Auth Middleware
	r.Use(func(c *gin.Context) {
		userIDStr := c.GetHeader("X-Test-User-ID")
		if userIDStr != "" {
			uid, _ := uuid.Parse(userIDStr)
			c.Set("user_id", uid)
		}
		c.Next()
	})

	r.POST("/documents/:id/share", share.InviteUsers)
	r.GET("/documents/:id/share", share.ListPermissions)
	r.DELETE("/documents/:id/share/:userId", share.RevokeAccess)
	
	r.POST("/documents/:id/request-access", share.RequestAccess)
	r.POST("/documents/:id/requests/:requestId", share.HandleRequest)
	r.GET("/access/requests", share.ListRequests)

	return r
}

func cleanupTestData(email string) {
	database.DB.Where("email = ?", email).Delete(&models.User{})
}

func TestSharePermissionFlow(t *testing.T) {
	godotenv.Load("../../.env")
	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	ownerEmail := "owner_share@example.com"
	invitedEmail := "invited_user@example.com"
	requesterEmail := "requester_user@example.com"

	cleanupTestData(ownerEmail)
	cleanupTestData(invitedEmail)
	cleanupTestData(requesterEmail)
	defer cleanupTestData(ownerEmail)
	defer cleanupTestData(invitedEmail)
	defer cleanupTestData(requesterEmail)

	// 1. Setup Users
	owner := models.User{Email: ownerEmail, Name: "Owner", PasswordHash: "hash"}
	invited := models.User{Email: invitedEmail, Name: "Invited", PasswordHash: "hash"}
	requester := models.User{Email: requesterEmail, Name: "Requester", PasswordHash: "hash"}
	
	database.DB.Create(&owner)
	database.DB.Create(&invited)
	database.DB.Create(&requester)

	// 2. Setup Document
	doc := models.Document{OwnerID: owner.ID, Title: "Permission Doc"}
	database.DB.Create(&doc)

	r := setupRouter()

	// --- TEST 1: Invite User ---
	t.Run("Invite User", func(t *testing.T) {
		reqBody := []map[string]interface{}{
			{
				"email": invitedEmail,
				"role":  "editor",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/documents/"+doc.ID.String()+"/share", bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}
		
		// Verify DB
		var perm models.DocumentPermission
		err := database.DB.Where("document_id = ? AND user_id = ?", doc.ID, invited.ID).First(&perm).Error
		if err != nil {
			t.Error("Permission not created in DB")
		}
		if perm.Role != models.RoleEditor {
			t.Errorf("Expected editor, got %s", perm.Role)
		}
	})

	// --- TEST 2: List Permissions ---
	t.Run("List Permissions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/documents/"+doc.ID.String()+"/share", nil)
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		if len(data) != 1 {
			t.Errorf("Expected 1 permission, got %d", len(data))
		}
	})

	// --- TEST 3: Request Access ---
	var requestID string
	t.Run("Request Access", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/documents/"+doc.ID.String()+"/request-access", nil)
		req.Header.Set("X-Test-User-ID", requester.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}
		
		var reqModel models.AccessRequest
		err := database.DB.Where("document_id = ? AND requester_id = ?", doc.ID, requester.ID).First(&reqModel).Error
		if err != nil {
			t.Error("Access Request not created in DB")
		}
		requestID = reqModel.ID.String()
	})

	// --- TEST 4: Handle Request (Approve) ---
	t.Run("Handle Request - Approve", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"action": "approved",
			"role":   "viewer",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/documents/"+doc.ID.String()+"/requests/"+requestID, bytes.NewBuffer(body))
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify Permission Created
		var perm models.DocumentPermission
		err := database.DB.Where("document_id = ? AND user_id = ?", doc.ID, requester.ID).First(&perm).Error
		if err != nil {
			t.Error("Permission not created after approval")
		}
		if perm.Role != models.RoleViewer {
			t.Errorf("Expected viewer, got %s", perm.Role)
		}
	})

	// --- TEST 5: Revoke Access ---
	t.Run("Revoke Access", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/documents/"+doc.ID.String()+"/share/"+invited.ID.String(), nil)
		req.Header.Set("X-Test-User-ID", owner.ID.String())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		// Verify Deleted
		var perm models.DocumentPermission
		err := database.DB.Where("document_id = ? AND user_id = ?", doc.ID, invited.ID).First(&perm).Error
		if err == nil {
			t.Error("Permission should be deleted")
		}
	})
}
