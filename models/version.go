package models

import (
	"time"

	"github.com/google/uuid"
)

type Version struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	VersionNumber   int       `gorm:"not null" json:"version_number"`
	ContentSnapshot string    `gorm:"type:text" json:"content_snapshot"`
	CreatedAt       time.Time `json:"created_at"`
}
