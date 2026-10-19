package share

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"github.com/personal-stakshi/draftly/backend/services"
)

func InitService() services.ShareService {
	permRepo := dbops.NewPermissionRepository()
	docRepo := dbops.NewDocumentRepository()
	userRepo := dbops.NewUserRepository()
	return services.NewShareService(permRepo, docRepo, userRepo)
}

func InviteUsers(c *gin.Context) {
	idStr := c.Param("id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid doc ID"})
		return
	}

	req, err := ValidateInviteRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	// validator uses one struct, service uses another, so map them here
	serviceInvites := make([]services.InviteRequestItem, len(*req))
	for i, item := range *req {
		serviceInvites[i] = services.InviteRequestItem{
			Email: item.Email,
			Role:  item.Role,
		}
	}

	invited, err := svc.InviteUsers(docID, userID, serviceInvites)
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Invitations processed", gin.H{"invited": invited})
}

func ListPermissions(c *gin.Context) {
	idStr := c.Param("id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid doc ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	perms, err := svc.ListPermissions(docID, userID)
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Permissions listed", TransformPermissions(perms))
}

func RevokeAccess(c *gin.Context) {
	docIDStr := c.Param("id")
	targetIDStr := c.Param("userId")
	
	docID, _ := uuid.Parse(docIDStr)
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid user ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	err = svc.RevokeAccess(docID, userID, targetID)
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Access revoked", nil)
}

func RequestAccess(c *gin.Context) {
	idStr := c.Param("id")
	docID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid doc ID"})
		return
	}

	input, err := ValidateRequestAccess(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	err = svc.RequestAccess(docID, userID, models.ShareRole(input.Role))
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Access requested", nil)
}

func HandleRequest(c *gin.Context) {
	reqIDStr := c.Param("requestId")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request ID"})
		return
	}

	input, err := ValidateHandleRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	err = svc.HandleAccessRequest(reqID, userID, models.RequestStatus(input.Action), models.ShareRole(input.Role))
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Request handled", nil)
}

func ListRequests(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	sent, received, err := svc.ListRequests(userID)
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	SendSuccessResponse(c, "Requests listed", gin.H{
		"sent":     TransformRequests(sent),
		"received": TransformRequests(received),
	})
}

func AccessByToken(c *gin.Context) {
	req, err := ValidateAccessTokenRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	svc := InitService()

	doc, err := svc.AccessByToken(req.Token, userID)
	if err != nil {
		SendErrorResponse(c, err)
		return
	}

	// send back the document info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":      doc.ID,
			"title":   doc.Title,
			"content": doc.Content,
			"owner_id": doc.OwnerID,
		},
	})
}

