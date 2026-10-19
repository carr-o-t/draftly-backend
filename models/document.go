package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	// max chars we accept for a document title
	MaxTitleLength = 255
	// since content is text typre so we cap it in the validators
	MaxContentLength = 1_000_000
)

type Document struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	Title     string    `gorm:"size:255;default:'Untitled Document'" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// relations
	Owner    User            `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`
	Shares   []DocumentShare `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"shares,omitempty"`
	Versions []Version       `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"versions,omitempty"`
}
