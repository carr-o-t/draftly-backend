package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func ValidateAuthRequest(c *gin.Context) (*AuthRequest, error) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	return &req, nil
}