package health

import "github.com/gin-gonic/gin"

func ValidateHealthCheckRequest(c *gin.Context) error {
	// nothing to validate here
	return nil
}
