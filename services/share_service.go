package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
)

type InviteRequestItem struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ShareService interface {
	InviteUsers(documentID uuid.UUID, ownerID uuid.UUID, invites []InviteRequestItem) ([]string, error)
	ListPermissions(documentID uuid.UUID, userID uuid.UUID) ([]models.DocumentPermission, error)
	RevokeAccess(documentID, ownerID, targetUserID uuid.UUID) error
	
	RequestAccess(documentID, requesterID uuid.UUID, role models.ShareRole) error
	HandleAccessRequest(requestID, ownerID uuid.UUID, action models.RequestStatus, role models.ShareRole) error
	ListRequests(userID uuid.UUID) (sent []models.AccessRequest, received []models.AccessRequest, err error)
	
	// share token access
	AccessByToken(token string, userID uuid.UUID) (*models.Document, error)
}

type shareService struct {
	permRepo  dbops.PermissionRepository
	docRepo   dbops.DocumentRepository
	userRepo  dbops.UserRepository
	shareRepo dbops.ShareRepository
}

func NewShareService(permRepo dbops.PermissionRepository, docRepo dbops.DocumentRepository, userRepo dbops.UserRepository) ShareService {
	return &shareService{
		permRepo:  permRepo,
		docRepo:   docRepo,
		userRepo:  userRepo,
		shareRepo: dbops.NewShareRepository(),
	}
}

func (s *shareService) InviteUsers(documentID uuid.UUID, ownerID uuid.UUID, invites []InviteRequestItem) ([]string, error) {
	// 1. only the owner can invite
	doc, err := s.docRepo.GetDocumentByID(documentID)
	if err != nil {
		return nil, err
	}
	if doc.OwnerID != ownerID {
		return nil, errors.New("permission denied: only owner can invite")
	}

	invitedEmails := []string{}

	// 2. go through each invite
	for _, invite := range invites {
		user, err := s.userRepo.GetUserByEmail(invite.Email)
		if err != nil {
			continue // no such user, skip it
		}

		if user.ID == ownerID {
			continue 
		}

		// role has to be viewer or editor
		if invite.Role != "viewer" && invite.Role != "editor" {
			continue // bad role, skip it
		}

		// swap out any old permission for this user
		_, err = s.permRepo.GetPermission(documentID, user.ID)
		if err == nil {
			s.permRepo.DeletePermission(documentID, user.ID)
		}

		perm := &models.DocumentPermission{
			DocumentID: documentID,
			UserID:     user.ID,
			Role:       models.ShareRole(invite.Role),
			CreatedAt:  time.Now(),
		}
		
		if err := s.permRepo.CreatePermission(perm); err == nil {
			invitedEmails = append(invitedEmails, invite.Email)
		}
	}

	return invitedEmails, nil
}

func (s *shareService) ListPermissions(documentID uuid.UUID, userID uuid.UUID) ([]models.DocumentPermission, error) {
	doc, err := s.docRepo.GetDocumentByID(documentID)
	if err != nil {
		return nil, err
	}
	if doc.OwnerID != userID {
		return nil, errors.New("permission denied")
	}

	return s.permRepo.ListPermissions(documentID)
}

func (s *shareService) RevokeAccess(documentID, ownerID, targetUserID uuid.UUID) error {
	doc, err := s.docRepo.GetDocumentByID(documentID)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return errors.New("permission denied")
	}

	return s.permRepo.DeletePermission(documentID, targetUserID)
}

func (s *shareService) RequestAccess(documentID, requesterID uuid.UUID, role models.ShareRole) error {
	// make shure the doc is there
	doc, err := s.docRepo.GetDocumentByID(documentID)
	if err != nil {
		return errors.New("document not found")
	}
	if doc.OwnerID == requesterID {
		return errors.New("cannot request access to own document")
	}

	// if they already have access, only let them ask for a higher role
	if perm, err := s.permRepo.GetPermission(documentID, requesterID); err == nil {
		if roleRank(perm.Role) >= roleRank(role) {
			return errors.New("already has access")
		}
		// they have a lower role (viewer) and want to go up to editor
	}

	// dont save a second pending request for the same doc and role
	existing, err := s.permRepo.ListAccessRequests(requesterID)
	if err == nil {
		for _, r := range existing {
			if r.DocumentID == documentID && r.Status == models.StatusPending && r.Role == role {
				return errors.New("request already pending")
			}
		}
	}

	// save the request
	req := &models.AccessRequest{
		DocumentID:  documentID,
		RequesterID: requesterID,
		Status:      models.StatusPending,
		Role:        role,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.permRepo.CreateAccessRequest(req)
}

// roleRank gives each role a number so we can compare who has more access
func roleRank(r models.ShareRole) int {
	if r == models.RoleEditor {
		return 2
	}
	return 1 // viewer
}

func (s *shareService) HandleAccessRequest(requestID, ownerID uuid.UUID, action models.RequestStatus, role models.ShareRole) error {
	req, err := s.permRepo.GetAccessRequest(requestID)
	if err != nil {
		return err
	}

	// make sure the owner making the call is the owner of this document
	doc, err := s.docRepo.GetDocumentByID(req.DocumentID)
	if err != nil {
		return err
	}
	if doc.OwnerID != ownerID {
		return errors.New("permission denied")
	}

	if req.Status != models.StatusPending {
		return errors.New("request already handled")
	}

	req.Status = action
	req.UpdatedAt = time.Now()

	if action == models.StatusApproved {
		// the owner role wins, otherwise use the role the requester asked for
		grantRole := role
		if grantRole == "" {
			grantRole = req.Role
		}
		if grantRole == "" {
			grantRole = models.RoleViewer
		}
		// give them the permission
		perm := &models.DocumentPermission{
			DocumentID: req.DocumentID,
			UserID:     req.RequesterID,
			Role:       grantRole,
			CreatedAt:  time.Now(),
		}
		if err := s.permRepo.CreatePermission(perm); err != nil {
			return err
		}
	}

	return s.permRepo.UpdateAccessRequest(req)
}

func (s *shareService) ListRequests(userID uuid.UUID) (sent []models.AccessRequest, received []models.AccessRequest, err error) {
	sent, err = s.permRepo.ListAccessRequests(userID)
	if err != nil {
		return nil, nil, err
	}

	received, err = s.permRepo.ListPendingRequestsForDocs(userID)
	if err != nil {
		return nil, nil, err
	}

	return sent, received, nil
}

func (s *shareService) AccessByToken(token string, userID uuid.UUID) (*models.Document, error) {
	// 1. check the token
	share, err := s.shareRepo.GetShareByToken(token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// 2. make sure it has not expired
	if time.Now().After(share.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	// 3. get the document
	doc, err := s.docRepo.GetDocumentByID(share.DocumentID)
	if err != nil {
		return nil, errors.New("document not found")
	}

	// 4. if the user is the owner or already has a permission, let them in
	if doc.OwnerID == userID {
		return doc, nil // owner always has access
	}

	_, err = s.permRepo.GetPermission(share.DocumentID, userID)
	if err == nil {
		return doc, nil // has permission already
	}

	// 5. no permission, so block them
	return nil, errors.New("access denied: please request access from the document owner")
}

