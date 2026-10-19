package dbops

import (
	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	// permissions
	CreatePermission(perm *models.DocumentPermission) error
	UpdatePermission(perm *models.DocumentPermission) error
	GetPermission(documentID, userID uuid.UUID) (*models.DocumentPermission, error)
	ListPermissions(documentID uuid.UUID) ([]models.DocumentPermission, error)
	DeletePermission(documentID, userID uuid.UUID) error
	ListSharedDocuments(userID uuid.UUID) ([]models.DocumentPermission, error)

	// access requests
	CreateAccessRequest(req *models.AccessRequest) error
	GetAccessRequest(id uuid.UUID) (*models.AccessRequest, error)
	UpdateAccessRequest(req *models.AccessRequest) error
	ListAccessRequests(userID uuid.UUID) ([]models.AccessRequest, error) // requests i sent
	ListPendingRequestsForDocs(ownerID uuid.UUID) ([]models.AccessRequest, error) // requests i recieve as owner
}

type postgresPermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository() PermissionRepository {
	return &postgresPermissionRepository{db: database.DB}
}

// --- permissions ---

func (r *postgresPermissionRepository) CreatePermission(perm *models.DocumentPermission) error {
	return r.db.Create(perm).Error
}

func (r *postgresPermissionRepository) UpdatePermission(perm *models.DocumentPermission) error {
	return r.db.Save(perm).Error
}

func (r *postgresPermissionRepository) GetPermission(documentID, userID uuid.UUID) (*models.DocumentPermission, error) {
	var perm models.DocumentPermission
	err := r.db.Where("document_id = ? AND user_id = ?", documentID, userID).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *postgresPermissionRepository) ListPermissions(documentID uuid.UUID) ([]models.DocumentPermission, error) {
	var perms []models.DocumentPermission
	err := r.db.Preload("User").Where("document_id = ?", documentID).Find(&perms).Error
	return perms, err
}

func (r *postgresPermissionRepository) DeletePermission(documentID, userID uuid.UUID) error {
	return r.db.Where("document_id = ? AND user_id = ?", documentID, userID).Delete(&models.DocumentPermission{}).Error
}

func (r *postgresPermissionRepository) ListSharedDocuments(userID uuid.UUID) ([]models.DocumentPermission, error) {
	var perms []models.DocumentPermission
	err := r.db.Preload("Document").Preload("Document.Owner").Where("user_id = ?", userID).Order("created_at desc").Find(&perms).Error
	return perms, err
}

// --- access requests ---

func (r *postgresPermissionRepository) CreateAccessRequest(req *models.AccessRequest) error {
	return r.db.Create(req).Error
}

func (r *postgresPermissionRepository) GetAccessRequest(id uuid.UUID) (*models.AccessRequest, error) {
	var req models.AccessRequest
	err := r.db.First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *postgresPermissionRepository) UpdateAccessRequest(req *models.AccessRequest) error {
	return r.db.Save(req).Error
}

func (r *postgresPermissionRepository) ListAccessRequests(userID uuid.UUID) ([]models.AccessRequest, error) {
	var reqs []models.AccessRequest
	err := r.db.Preload("Document").Where("requester_id = ?", userID).Order("created_at desc").Find(&reqs).Error
	return reqs, err
}

func (r *postgresPermissionRepository) ListPendingRequestsForDocs(ownerID uuid.UUID) ([]models.AccessRequest, error) {
	// find requests where the document owner is ownerID
	var reqs []models.AccessRequest
	// join documents so we can filter on the owner
	err := r.db.Joins("Document").
		Preload("Requester").
		// Where("status = ?", models.StatusPending).
		Where("\"Document\".owner_id = ?", ownerID). // table alias is case sensitive here
		Order("created_at desc").
		Find(&reqs).Error
	return reqs, err
}
