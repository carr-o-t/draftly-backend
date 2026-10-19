package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/controllers/auth"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
)

// set up the router for the tests
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/v1/auth/login-or-register", auth.LoginOrRegister)
	return r
}

// clean up test data
func cleanupTestData(email, ip string) {
	database.DB.Where("email = ?", email).Delete(&models.User{})
	database.DB.Where("ip_address = ?", ip).Delete(&models.IPUsage{})
}

func TestAuthFlow(t *testing.T) {
	// init config and db
	godotenv.Load("../../.env")

	cfg := config.LoadConfig()
	database.ConnectDatabase(cfg)

	testEmail := "test_auto@example.com"
	testIP := "10.0.0.99"

	// start and end clean
	cleanupTestData(testEmail, testIP)
	defer cleanupTestData(testEmail, testIP)

	r := setupRouter()

	t.Run("Positive: Register New User", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/v1/auth/login-or-register", bytes.NewBuffer(body))
		req.RemoteAddr = testIP + ":1234"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", testIP) // fake the ip
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp["success"] != true {
			t.Error("Expected success true")
		}

		data := resp["data"].(map[string]interface{})
		if data["is_new"] != true {
			t.Error("Expected is_new true")
		}
	})

	t.Run("Positive: Login Existing User", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    testEmail,
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/v1/auth/login-or-register", bytes.NewBuffer(body))
		req.RemoteAddr = testIP + ":1234"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", testIP)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp["data"].(map[string]interface{})
		if data["is_new"] != false {
			t.Error("Expected is_new false")
		}
	})

	t.Run("Negative: Wrong Password", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    testEmail,
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/v1/auth/login-or-register", bytes.NewBuffer(body))
		req.RemoteAddr = testIP + ":1234"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Negative: Invalid Email Format", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "not-an-email",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/v1/auth/login-or-register", bytes.NewBuffer(body))
		req.RemoteAddr = testIP + ":1234"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}
