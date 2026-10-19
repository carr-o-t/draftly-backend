package share

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
)

type PermissionResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type AccessRequestResponse struct {
	ID        string `json:"id"`
	DocID     string `json:"document_id"`
	DocTitle  string `json:"doc_title,omitempty"`
	RequesterID string `json:"requester_id"`
	RequesterEmail string `json:"requester_email"`
	Requester string `json:"requester_name,omitempty"`
	Status    string `json:"status"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SendSuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendErrorResponse(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if err != nil {
		msg := err.Error()
		if msg == "permission denied" || msg == "permission denied: only owner can invite" {
			status = http.StatusForbidden
		} else if msg == "access denied: please request access from the document owner" {
			status = http.StatusUnauthorized
		} else if msg == "document not found" || msg == "request not found" {
			status = http.StatusNotFound
		} else if msg == "already has access" || msg == "request already pending" {
			status = http.StatusConflict
		} else if msg == "invalid or expired token" {
			status = http.StatusUnauthorized
		}
		
		c.JSON(status, StandardResponse{
			Success: false,
			Message: msg,
		})
		return
	}
	c.JSON(status, StandardResponse{Success: false, Message: "Unknown error"})
}

func TransformPermissions(perms []models.DocumentPermission) []PermissionResponse {
	resp := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		resp[i] = PermissionResponse{
			UserID: p.UserID.String(),
			Email:  p.User.Email, // loaded with preload
			Name:   p.User.Name,
			Role:   string(p.Role),
		}
	}
	return resp
}

func TransformRequests(reqs []models.AccessRequest) []AccessRequestResponse {
	resp := make([]AccessRequestResponse, len(reqs))
	for i, r := range reqs {
		resp[i] = AccessRequestResponse{
			ID:          r.ID.String(),
			DocID:       r.DocumentID.String(),
			DocTitle:    r.Document.Title, // loaded with preload
			RequesterID: r.RequesterID.String(),
			RequesterEmail: r.Requester.Email, // loaded with preload
			Requester:   r.Requester.Name, // loaded with preload
			Status:      string(r.Status),
			Role:      string(r.Role),
			CreatedAt: r.CreatedAt.String(),
		}
	}
	return resp
}
