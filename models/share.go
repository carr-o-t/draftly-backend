package models

import (
	"time"

	"github.com/google/uuid"
)

type ShareRole string

const (
	RoleViewer ShareRole = "viewer"
	RoleEditor ShareRole = "editor"
)

type DocumentShare struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID  uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	Token       string    `gorm:"uniqueIndex;not null" json:"token"`
	// RoleGranted removed as per new hybrid flow (Token is just entry point)
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}
