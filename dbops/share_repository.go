package dbops

import (
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

type ShareRepository interface {
	CreateShare(share *models.DocumentShare) error
	GetShareByToken(token string) (*models.DocumentShare, error)
	DeleteShare(documentID uuid.UUID) error // Revoke all for doc? or by token? Assuming one active link for MVP simplicity
	GetActiveShare(documentID uuid.UUID) (*models.DocumentShare, error)
}

type postgresShareRepository struct {
	db *gorm.DB
}

func NewShareRepository() ShareRepository {
	return &postgresShareRepository{db: database.DB}
}

func (r *postgresShareRepository) CreateShare(share *models.DocumentShare) error {
	return r.db.Create(share).Error
}

func (r *postgresShareRepository) GetShareByToken(token string) (*models.DocumentShare, error) {
	var share models.DocumentShare
	result := r.db.Where("token = ?", token).First(&share)
	if result.Error != nil {
		return nil, result.Error
	}
	return &share, nil
}

func (r *postgresShareRepository) GetActiveShare(documentID uuid.UUID) (*models.DocumentShare, error) {
	var share models.DocumentShare
	result := r.db.Where("document_id = ?", documentID).First(&share)
	if result.Error != nil {
		return nil, result.Error
	}
	return &share, nil
}

func (r *postgresShareRepository) DeleteShare(documentID uuid.UUID) error {
	return r.db.Where("document_id = ?", documentID).Delete(&models.DocumentShare{}).Error
}
