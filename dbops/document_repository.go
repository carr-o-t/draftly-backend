package dbops

import (
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	CreateDocument(doc *models.Document) error
	GetDocumentByID(id uuid.UUID) (*models.Document, error)
	ListDocumentsByOwner(ownerID uuid.UUID) ([]models.Document, error)
	CountDocumentsByOwner(ownerID uuid.UUID) (int64, error)
	UpdateDocument(doc *models.Document) error
	DeleteDocument(id uuid.UUID) error
	// used when we need the shares along with the doc
	GetDocumentWithShares(id uuid.UUID) (*models.Document, error)
}

type postgresDocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository() DocumentRepository {
	return &postgresDocumentRepository{db: database.DB}
}

func (r *postgresDocumentRepository) CreateDocument(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *postgresDocumentRepository) GetDocumentByID(id uuid.UUID) (*models.Document, error) {
	var doc models.Document
	result := r.db.First(&doc, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}

func (r *postgresDocumentRepository) ListDocumentsByOwner(ownerID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	result := r.db.Where("owner_id = ?", ownerID).Order("updated_at desc").Find(&docs)
	return docs, result.Error
}

func (r *postgresDocumentRepository) CountDocumentsByOwner(ownerID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.Model(&models.Document{}).Where("owner_id = ?", ownerID).Count(&count)
	return count, result.Error
}

func (r *postgresDocumentRepository) UpdateDocument(doc *models.Document) error {
	return r.db.Save(doc).Error
}

func (r *postgresDocumentRepository) DeleteDocument(id uuid.UUID) error {
	return r.db.Delete(&models.Document{}, "id = ?", id).Error
}

func (r *postgresDocumentRepository) GetDocumentWithShares(id uuid.UUID) (*models.Document, error) {
	var doc models.Document
	result := r.db.Preload("Shares").First(&doc, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &doc, nil
}
