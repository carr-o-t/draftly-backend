package app

import (
	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/auth"
	"github.com/personal-stakshi/draftly/backend/controllers/health"
)

func SetupRoutes(r *gin.Engine) {
	// Global Middleware
	// TODO: Implement CORS
	// r.Use(CorsMiddleware()) 

	// Health Check
	r.GET("/health", health.HealthCheck)

	// API V1 Group
	v1 := r.Group("/v1")
	{
		// Auth Routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login-or-register", auth.LoginOrRegister)
		}
	}
}
