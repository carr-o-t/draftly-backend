package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
)

type AuthResponseData struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
	IsNew bool        `json:"is_new"`
}

type StandardResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    *AuthResponseData `json:"data,omitempty"`
}

func SendAuthResponse(c *gin.Context, user *models.User, token string, isNew bool, err error) {
	if err != nil {
		// map errors to http status codes
		status := http.StatusInternalServerError
		
		switch err.Error() {
		case "invalid credentials":
			status = http.StatusUnauthorized
		case "registration limit exceeded for this IP":
			status = http.StatusTooManyRequests
		case "failed to hash password", "failed to create user", "failed to generate token":
			status = http.StatusInternalServerError
		default:
			// unknown errors just fall back to 500
			status = http.StatusInternalServerError
		}

		c.JSON(status, StandardResponse{
			Success: false, 
			Message: err.Error(),
		})
		return
	}

	// all good
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: &AuthResponseData{
			Token: token,
			User:  *user,
			IsNew: isNew,
		},
	})
}