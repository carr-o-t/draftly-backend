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
	GetDocument(id uuid.UUID, userID uuid.UUID) (*models.Document, string, error) // Returns doc, role, error
	ListDocuments(userID uuid.UUID) ([]models.Document, []models.Document, error) // Returns myDocs, sharedDocs, error
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

	// Auto-generate Share Token
	token := generateSecureToken()
	
	share := &models.DocumentShare{
		DocumentID: doc.ID,
		Token:      token,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * 365 * 10 * time.Hour), // 10 Years
	}
	// Best effort creation of share token
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

	// 1. Check Owner
	if doc.OwnerID == userID {
		return doc, "owner", nil
	}

	// 2. Check Permissions (Direct Invite)
	perm, err := s.permRepo.GetPermission(id, userID)
	if err == nil {
		return doc, string(perm.Role), nil
	}

	// 3. Fallback: Check Token? 
	// The current requirement and GetDocument signature relies on userID.
	// If the user accessed via a public link (token), the Client normally hits the "Access Share Link" API first/separately.
	// However, if the user is logged in, and clicking a link, maybe we should auto-grant permission?
	// For "GetDocument", strict check: Owner OR Permission. 
	// Access via Token is handled by `CreatePermission` or separate flow?
	// Phase 4 "Access Token" logic was: Return doc info. 
	// If the link is just an entry point, users might still need to specific "Request Access" if they are not invited.
	// BUT, if the link is a "Magic Link" that gives access... DocumentShare has NO ROLE now.
	// So purely clicking the link does NOT grant access unless the app logic says so.
	// The user said: "There is to be only one share token... remove role granted". 
	// This implies the token is just for LOCATING the document or Starting the flow.
	// So Access Denied here is correct if no explicit permission exists.

	return nil, "", errors.New("access denied")
}

func (s *documentService) ListDocuments(userID uuid.UUID) ([]models.Document, []models.Document, error) {
	// my_documents
	myDocs, err := s.docRepo.ListDocumentsByOwner(userID)
	if err != nil {
		return nil, nil, err
	}

	// shared_with_me (From Permissions)
	perms, err := s.permRepo.ListSharedDocuments(userID)
	if err != nil {
		return nil, nil, err
	}
	
	sharedDocs := make([]models.Document, len(perms))
	for i, p := range perms {
		// p.Document is preloaded
		sharedDocs[i] = p.Document
		// Note: p.Document.Owner is also preloaded if `ListSharedDocuments` did it.
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
	doc.Content = content
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
	
	// Autosave only updates content
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
