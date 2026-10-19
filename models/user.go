package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	MaxEmailLength    = 255
	MaxNameLength     = 255
	MaxPasswordLength = 72 // bc bcryt only reads first 72 bytes
	MaxIPLength       = 45 // max len of ipv6 string
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"`
	Name         string    `gorm:"size:255" json:"name"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// relations
	Documents []Document `gorm:"foreignKey:OwnerID" json:"documents,omitempty"`
}
