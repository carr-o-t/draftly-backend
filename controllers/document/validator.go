package document

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/models"
)

func tooLong(field string, value string, max int) error {
	if utf8.RuneCountInString(value) > max {
		return fmt.Errorf("%s must be at most %d characters", field, max)
	}
	return nil
}

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
	if err := tooLong("title", req.Title, models.MaxTitleLength); err != nil {
		return nil, err
	}
	return &req, nil
}

func ValidateUpdateDocumentRequest(c *gin.Context) (*UpdateDocumentRequest, error) {
	var req UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	if err := tooLong("title", req.Title, models.MaxTitleLength); err != nil {
		return nil, err
	}
	if err := tooLong("content", req.Content, models.MaxContentLength); err != nil {
		return nil, err
	}
	return &req, nil
}

func ValidateAutosaveRequest(c *gin.Context) (*AutosaveRequest, error) {
	var req AutosaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, errors.New(err.Error())
	}
	if err := tooLong("content", req.Content, models.MaxContentLength); err != nil {
		return nil, err
	}
	return &req, nil
}
