package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	Title     string    `gorm:"default:'Untitled Document'" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Owner    User            `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Shares   []DocumentShare `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"shares,omitempty"`
	Versions []Version       `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"versions,omitempty"`
}
