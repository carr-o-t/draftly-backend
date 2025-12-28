package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid authorization header format"})
			c.Abort()
			return
		}

		cfg := config.LoadConfig()
		claims, err := utils.ValidateToken(parts[1], cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
