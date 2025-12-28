package models

import (
	"time"

	"github.com/google/uuid"
)

type IPUsage struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IPAddress       string    `gorm:"uniqueIndex;not null" json:"ip_address"`
	CreatedAccounts int       `gorm:"default:0" json:"created_accounts"`
	LastAttemptAt   time.Time `json:"last_attempt_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
