package models

import (
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
)

type AccessRequest struct {
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID  uuid.UUID     `gorm:"type:uuid;not null;index" json:"document_id"`
	RequesterID uuid.UUID     `gorm:"type:uuid;not null;index" json:"requester_id"`
	Role        ShareRole     `gorm:"type:varchar(20);default:'viewer'" json:"role"`
	Status      RequestStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// relations
	Document  Document `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"document,omitempty"`
	Requester User     `gorm:"foreignKey:RequesterID;constraint:OnDelete:CASCADE" json:"requester,omitempty"`
}

type DocumentPermission struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Role       ShareRole `gorm:"type:varchar(20);not null" json:"role"`
	CreatedAt  time.Time `json:"created_at"`

	// relations
	Document Document `gorm:"foreignKey:DocumentID;constraint:OnDelete:CASCADE" json:"document,omitempty"`
	User     User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}
