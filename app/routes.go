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
	// todo: add cors

	r.GET("/health", health.HealthCheck)

	v1 := r.Group("/v1")
	{
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

		v1.GET("/access/requests", middleware.AuthMiddleware(), share.ListRequests)
		v1.POST("/share/access", middleware.AuthMiddleware(), share.AccessByToken)
	}

	// websocket route for collab. the token is checked inside ServeWs
	// (browsers send cookies with same-domain ws requests anyway)
	r.GET("/ws/:id", func(c *gin.Context) {
		realtime.ServeWs(hub, c)
	})
}
