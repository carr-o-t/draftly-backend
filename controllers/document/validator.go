package document

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type CreateDocumentRequest struct {
	Title string `json:"title"` // not required
}

type UpdateDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AutosaveRequest struct {
	Content string `json:"content" binding:"required"`
}

func ValidateCreateDocumentRequest(c *gin.Context) (*CreateDocumentRequest, error) {
	var req CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	return &req, nil
}

func ValidateUpdateDocumentRequest(c *gin.Context) (*UpdateDocumentRequest, error) {
	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	return &req, nil
}

func ValidateAutosaveRequest(c *gin.Context) (*AutosaveRequest, error) {
	var req AutosaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	return &req, nil
}
