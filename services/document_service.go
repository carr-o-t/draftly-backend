package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

type DocumentService interface {
	CreateDocument(ownerID uuid.UUID, title string) (*models.Document, error)
	GetDocument(id uuid.UUID, userID uuid.UUID) (*models.Document, string, error) // returns doc, role, error
	ListDocuments(userID uuid.UUID) ([]models.Document, []models.Document, error) // returns my docs, shared docs, error
	UpdateDocument(id uuid.UUID, userID uuid.UUID, title, content string) (*models.Document, error)
	DeleteDocument(id uuid.UUID, userID uuid.UUID) error
	AutoSave(id uuid.UUID, userID uuid.UUID, content string) (*models.Document, error)
}

type documentService struct {
	docRepo   dbops.DocumentRepository
	shareRepo dbops.ShareRepository
	permRepo  dbops.PermissionRepository
}

func NewDocumentService(repo dbops.DocumentRepository, shareRepo dbops.ShareRepository, permRepo dbops.PermissionRepository) DocumentService {
	return &documentService{
		docRepo:   repo,
		shareRepo: shareRepo,
		permRepo:  permRepo,
	}
}

func (s *documentService) CreateDocument(ownerID uuid.UUID, title string) (*models.Document, error) {
	if title == "" {
		title = "Untitled Document"
	}
	doc := &models.Document{
		OwnerID: ownerID,
		Title:   title,
		Content: "",
	}
	if err := s.docRepo.CreateDocument(doc); err != nil {
		return nil, err
	}

	// make a share token
	token := generateSecureToken()
	
	share := &models.DocumentShare{
		DocumentID: doc.ID,
		Token:      token,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * 365 * 10 * time.Hour), // 10 years
	}
	// best effort, not fatal if it fails
	s.shareRepo.CreateShare(share)

	return doc, nil
}

func generateSecureToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *documentService) GetDocument(id uuid.UUID, userID uuid.UUID) (*models.Document, string, error) {
	doc, err := s.docRepo.GetDocumentWithShares(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("document not found")
		}
		return nil, "", err
	}

	// 1. owner check
	if doc.OwnerID == userID {
		return doc, "owner", nil
	}

	// 2. direct invite check
	perm, err := s.permRepo.GetPermission(id, userID)
	if err == nil {
		return doc, string(perm.Role), nil
	}

	// 3. no owner or permission found.
	// the share token just locates the document, it does not grant any role.
	// users need an explicit permission here, otherwise they have to request access.
	return nil, "", errors.New("access denied")
}

func (s *documentService) ListDocuments(userID uuid.UUID) ([]models.Document, []models.Document, error) {
	// docs i own
	myDocs, err := s.docRepo.ListDocumentsByOwner(userID)
	if err != nil {
		return nil, nil, err
	}

	// docs shared with me
	perms, err := s.permRepo.ListSharedDocuments(userID)
	if err != nil {
		return nil, nil, err
	}
	
	sharedDocs := make([]models.Document, len(perms))
	for i, p := range perms {
		// document is preloaded by ListSharedDocuments
		sharedDocs[i] = p.Document
	}

	return myDocs, sharedDocs, nil
}

func (s *documentService) UpdateDocument(id uuid.UUID, userID uuid.UUID, title, content string) (*models.Document, error) {
	doc, role, err := s.GetDocument(id, userID)
	if err != nil {
		return nil, err
	}
	
	if role != "owner" && role != "editor" {
		return nil, errors.New("permission denied")
	}

	doc.Title = title
	// content comes from the yjs doc via autosave. only overwrite the saved
	// html snapshot when a non-empty value is given, so a plain rename
	// never wipes it.
	if content != "" {
		doc.Content = content
	}
	doc.UpdatedAt = time.Now()

	if err := s.docRepo.UpdateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *documentService) AutoSave(id uuid.UUID, userID uuid.UUID, content string) (*models.Document, error) {
	doc, role, err := s.GetDocument(id, userID)
	if err != nil {
		return nil, err
	}

	if role != "owner" && role != "editor" {
		return nil, errors.New("permission denied")
	}
	
	// autosave only touches content
	doc.Content = content
	doc.UpdatedAt = time.Now()
	
	if err := s.docRepo.UpdateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *documentService) DeleteDocument(id uuid.UUID, userID uuid.UUID) error {
	doc, err := s.docRepo.GetDocumentByID(id)
	if err != nil {
		return err
	}

	if doc.OwnerID != userID {
		return errors.New("permission denied: only owner can delete")
	}

	return s.docRepo.DeleteDocument(id)
}
