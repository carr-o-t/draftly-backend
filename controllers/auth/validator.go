package auth

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
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
	if utf8.RuneCountInString(req.Email) > models.MaxEmailLength {
		return nil, fmt.Errorf("email must be at most %d characters", models.MaxEmailLength)
	}
	if utf8.RuneCountInString(req.Password) > models.MaxPasswordLength {
		return nil, fmt.Errorf("password must be at most %d characters", models.MaxPasswordLength)
	}
	return &req, nil
}