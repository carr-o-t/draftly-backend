package document

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
	"github.com/personal-stakshi/draftly/backend/sanitize"
	"github.com/personal-stakshi/draftly/backend/services"
)

type DocumentResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Role      string `json:"role,omitempty"` // owner, editor or viewer
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListDocumentsResponse struct {
	MyDocuments     []DocumentResponse `json:"my_documents"`
	SharedDocuments []DocumentResponse `json:"shared_with_me"`
}

type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func transformDocument(doc *models.Document, role string) DocumentResponse {
	return DocumentResponse{
		ID:        doc.ID.String(),
		Title:     doc.Title,
		// also clean rows saved before we sanitized on write
		Content:   sanitize.HTML(doc.Content),
		Role:      role,
		CreatedAt: doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt: doc.UpdatedAt.Format(time.RFC3339),
	}
}

func SendDocumentResponse(c *gin.Context, doc *models.Document, role string, err error) {
	if err != nil {
		status := http.StatusInternalServerError
		errStr := err.Error()
		if errStr == "document not found" {
			status = http.StatusNotFound
		} else if errStr == services.ErrDocumentLimitReached.Error() {
			status = http.StatusForbidden
		} else if strings.Contains(errStr, "access denied") || strings.Contains(errStr, "permission denied") {
			status = http.StatusForbidden
		} else {
			status = http.StatusInternalServerError
		}

		c.JSON(status, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    transformDocument(doc, role),
	})
}

func SendListDocumentsResponse(c *gin.Context, myDocs []models.Document, sharedDocs []models.Document, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	myDocsResp := make([]DocumentResponse, len(myDocs))
	for i, d := range myDocs {
		myDocsResp[i] = transformDocument(&d, "owner")
	}

	sharedDocsResp := make([]DocumentResponse, len(sharedDocs))
	for i, d := range sharedDocs {
		// todo: fetch teh real role from the share entry
		sharedDocsResp[i] = transformDocument(&d, "viewer") 
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: ListDocumentsResponse{
			MyDocuments:     myDocsResp,
			SharedDocuments: sharedDocsResp,
		},
	})
}

func SendSuccessResponse(c *gin.Context, message string, err error) {
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "document not found":
			status = http.StatusNotFound
		case "permission denied: only owner can delete":
			status = http.StatusForbidden
		}
		c.JSON(status, StandardResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, StandardResponse{Success: true, Message: message})
}
