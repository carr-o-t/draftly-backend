package realtime

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/personal-stakshi/draftly/backend/models"
)

// resolveRole returns the role (owner, editor, viewer) a user has on a
// document, or an error if they have no access at all.
func resolveRole(db *gorm.DB, docID, userID uuid.UUID) (string, error) {
	var doc models.Document
	if err := db.First(&doc, "id = ?", docID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("document not found")
		}
		return "", err
	}

	if doc.OwnerID == userID {
		return "owner", nil
	}

	var perm models.DocumentPermission
	err := db.Where("document_id = ? AND user_id = ?", docID, userID).First(&perm).Error
	if err == nil {
		return string(perm.Role), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return "", errors.New("access denied")
}
