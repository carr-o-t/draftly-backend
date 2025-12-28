package app

import (
	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/auth"
	"github.com/personal-stakshi/draftly/backend/controllers/document"
	"github.com/personal-stakshi/draftly/backend/controllers/health"
	"github.com/personal-stakshi/draftly/backend/middleware"
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

		// Documents Routes (Protected)
		docsGroup := v1.Group("/documents")
		docsGroup.Use(middleware.AuthMiddleware())
		{
			docsGroup.POST("", document.CreateDocument)
			docsGroup.GET("", document.ListDocuments)
			docsGroup.GET("/:id", document.GetDocument)
			docsGroup.PATCH("/:id", document.UpdateDocument)
			docsGroup.DELETE("/:id", document.DeleteDocument)
			docsGroup.PATCH("/:id/autosave", document.AutoSave)
		}
	}
}
