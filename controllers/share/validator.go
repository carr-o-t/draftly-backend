package share

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
)

type InviteRequestItem struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=viewer editor"`
}

type InviteUserRequest []InviteRequestItem // Array root

func ValidateInviteRequest(c *gin.Context) (*InviteUserRequest, error) {
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	if len(req) == 0 {
		return nil, errors.New("at least one invite is required")
	}
	return &req, nil
}

type HandleRequestInput struct {
	Action string `json:"action" binding:"required,oneof=approved rejected"`
	Role   string `json:"role"` // Optional if rejected, required if approved
}

func ValidateHandleRequest(c *gin.Context) (*HandleRequestInput, error) {
	var req HandleRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	if req.Action == string(models.StatusApproved) {
		if req.Role != "viewer" && req.Role != "editor" {
			return nil, errors.New("valid role (viewer/editor) is required for approval")
		}
	}
	return &req, nil
}

type AccessTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func ValidateAccessTokenRequest(c *gin.Context) (*AccessTokenRequest, error) {
	var req AccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

type RequestAccessInput struct {
	Role string `json:"role" binding:"required,oneof=viewer editor"`
}

func ValidateRequestAccess(c *gin.Context) (*RequestAccessInput, error) {
	var req RequestAccessInput
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
