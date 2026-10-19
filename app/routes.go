package app

import (
	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/controllers/auth"
	"github.com/personal-stakshi/draftly/backend/controllers/document"
	"github.com/personal-stakshi/draftly/backend/controllers/health"
	"github.com/personal-stakshi/draftly/backend/controllers/share"
	"github.com/personal-stakshi/draftly/backend/middleware"
	"github.com/personal-stakshi/draftly/backend/realtime"
)

func SetupRoutes(r *gin.Engine, hub *realtime.Hub) {
	// global middleware
	// todo: add cors
	// r.Use(CorsMiddleware())

	// health check
	r.GET("/health", health.HealthCheck)

	// all v1 api routes
	v1 := r.Group("/v1")
	{
		// auth
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login-or-register", auth.LoginOrRegister)
		}

		// documents (needs login)
		docsGroup := v1.Group("/documents")
		docsGroup.Use(middleware.AuthMiddleware())
		{
			docsGroup.POST("", document.CreateDocument)
			docsGroup.GET("", document.ListDocuments)
			docsGroup.GET("/:id", document.GetDocument)
			docsGroup.PATCH("/:id", document.UpdateDocument)
			docsGroup.DELETE("/:id", document.DeleteDocument)
			docsGroup.PATCH("/:id/autosave", document.AutoSave)
			
			// sharing and permissions
			docsGroup.POST("/:id/share", share.InviteUsers)
			docsGroup.GET("/:id/share", share.ListPermissions)
			docsGroup.DELETE("/:id/share/:userId", share.RevokeAccess)
			
			docsGroup.POST("/:id/request-access", share.RequestAccess)
			docsGroup.POST("/:id/requests/:requestId", share.HandleRequest)
		}
		
		// list of access requests
		v1.GET("/access/requests", middleware.AuthMiddleware(), share.ListRequests)
		

		// access a doc through a share token
		v1.POST("/share/access", middleware.AuthMiddleware(), share.AccessByToken)
	}

	// websocket route for collab. the token is checked inside ServeWs
	// (browsers send cookies with same-domain ws requests anyway)
	r.GET("/ws/:id", func(c *gin.Context) {
		realtime.ServeWs(hub, c)
	})
}
