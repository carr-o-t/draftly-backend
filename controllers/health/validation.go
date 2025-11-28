package health

import "github.com/gin-gonic/gin"

func ValidateHealthCheckRequest(c *gin.Context) error {
	// No validation needed for health check
	return nil
}
