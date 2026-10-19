package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentUpdate stores the yjs binary update blobs
type DocumentUpdate struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	Data       []byte    `gorm:"type:bytea" json:"data"`
	CreatedAt  time.Time `json:"created_at"`

	// relations
	Document Document `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"-"`
}
