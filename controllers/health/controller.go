package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/services"
)

func HealthCheck(c *gin.Context) {
	// Validation 
	if err := ValidateHealthCheckRequest(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Service Call
	service := services.NewHealthService()
	message := service.Check()

	// Transformation
	response := TransformHealthCheckResponse(message)

	c.JSON(http.StatusOK, response)
}
