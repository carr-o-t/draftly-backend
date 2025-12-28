package document

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/services"
)

func CreateDocument(c *gin.Context) {
	req, err := ValidateCreateDocumentRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	
	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	doc, err := svc.CreateDocument(userID, req.Title)
	SendDocumentResponse(c, doc, "owner", err)
}

func GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid document ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	doc, role, err := svc.GetDocument(id, userID)
	SendDocumentResponse(c, doc, role, err)
}

func ListDocuments(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	myDocs, sharedDocs, err := svc.ListDocuments(userID)
	SendListDocumentsResponse(c, myDocs, sharedDocs, err)
}

func UpdateDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid document ID"})
		return
	}

	req, err := ValidateUpdateDocumentRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	doc, err := svc.UpdateDocument(id, userID, req.Title, req.Content)
	// Role is implicitly confirmed as having edit rights if successful
	// But we don't know exact role here without querying again or modifying service return
	// For API response it's fine to just return the doc.
	SendDocumentResponse(c, doc, "", err) 
}

func DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid document ID"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	err = svc.DeleteDocument(id, userID)
	SendSuccessResponse(c, "Document deleted successfully", err)
}

func AutoSave(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid document ID"})
		return
	}

	req, err := ValidateAutosaveRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	repo := dbops.NewDocumentRepository()
	svc := services.NewDocumentService(repo)

	doc, err := svc.AutoSave(id, userID, req.Content)
	SendDocumentResponse(c, doc, "", err)
}
