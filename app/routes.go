package app

import (
	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/health"
)

func SetupRoutes(r *gin.Engine) {
	// Global Middleware
	// r.Use(CorsMiddleware()) // TODO: Implement CORS

	// Health Check
	r.GET("/health", health.HealthCheck)

	// API Group
	// api := r.Group("/api")
	// {
		// Define other routes here
	// }
}
