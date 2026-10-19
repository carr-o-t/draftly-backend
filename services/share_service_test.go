package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

func newShareSvc(p *fakePermRepo, d *fakeDocRepo, u *fakeUserRepo, s *fakeShareRepo) ShareService {
	if u == nil {
		u = &fakeUserRepo{}
	}
	if s == nil {
		s = &fakeShareRepo{}
	}
	return &shareService{permRepo: p, docRepo: d, userRepo: u, shareRepo: s}
}

func TestInviteUsers(t *testing.T) {
	owner := uuid.New()
	invitee := uuid.New()
	docID := uuid.New()

	t.Run("invites an existing user", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		userRepo := &fakeUserRepo{user: &models.User{ID: invitee, Email: "a@b.com"}}
		permRepo := &fakePermRepo{permErr: gorm.ErrRecordNotFound}

		invited, err := newShareSvc(permRepo, docRepo, userRepo, nil).
			InviteUsers(docID, owner, []InviteRequestItem{{Email: "a@b.com", Role: "editor"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invited) != 1 || invited[0] != "a@b.com" {
			t.Fatalf("wrong invited list: %v", invited)
		}
		if permRepo.createdPerm == nil || permRepo.createdPerm.Role != models.RoleEditor {
			t.Fatal("permission was not created with the editor role")
		}
	})

	t.Run("replaces an existing permission", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		userRepo := &fakeUserRepo{user: &models.User{ID: invitee, Email: "a@b.com"}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleViewer}}

		if _, err := newShareSvc(permRepo, docRepo, userRepo, nil).
			InviteUsers(docID, owner, []InviteRequestItem{{Email: "a@b.com", Role: "editor"}}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if permRepo.deletedUser != invitee {
			t.Fatal("old permission should have been deleted first")
		}
	})

	t.Run("only the owner can invite", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		_, err := newShareSvc(&fakePermRepo{}, docRepo, nil, nil).
			InviteUsers(docID, uuid.New(), []InviteRequestItem{{Email: "a@b.com", Role: "editor"}})
		if err == nil {
			t.Fatal("expected a permission error")
		}
	})

	t.Run("skips unknown emails, owners and bad roles", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}

		unknown := &fakeUserRepo{userErr: gorm.ErrRecordNotFound}
		if invited, _ := newShareSvc(&fakePermRepo{}, docRepo, unknown, nil).
			InviteUsers(docID, owner, []InviteRequestItem{{Email: "no@one.com", Role: "editor"}}); len(invited) != 0 {
			t.Fatalf("unknown email should be skipped, got %v", invited)
		}

		self := &fakeUserRepo{user: &models.User{ID: owner, Email: "me@b.com"}}
		if invited, _ := newShareSvc(&fakePermRepo{}, docRepo, self, nil).
			InviteUsers(docID, owner, []InviteRequestItem{{Email: "me@b.com", Role: "editor"}}); len(invited) != 0 {
			t.Fatalf("owner should be skipped, got %v", invited)
		}

		permRepo := &fakePermRepo{permErr: gorm.ErrRecordNotFound}
		badRole := &fakeUserRepo{user: &models.User{ID: invitee, Email: "a@b.com"}}
		if invited, _ := newShareSvc(permRepo, docRepo, badRole, nil).
			InviteUsers(docID, owner, []InviteRequestItem{{Email: "a@b.com", Role: "admin"}}); len(invited) != 0 {
			t.Fatalf("bad role should be skipped, got %v", invited)
		}
	})
}

func TestListPermissions(t *testing.T) {
	owner := uuid.New()
	docID := uuid.New()

	t.Run("owner sees permissions", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{listPerms: []models.DocumentPermission{{Role: models.RoleViewer}}}

		perms, err := newShareSvc(permRepo, docRepo, nil, nil).ListPermissions(docID, owner)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(perms) != 1 {
			t.Fatalf("expected 1 permission, got %d", len(perms))
		}
	})

	t.Run("non owner is rejected", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if _, err := newShareSvc(&fakePermRepo{}, docRepo, nil, nil).ListPermissions(docID, uuid.New()); err == nil {
			t.Fatal("expected a permission error")
		}
	})
}

func TestRevokeAccess(t *testing.T) {
	owner := uuid.New()
	docID := uuid.New()

	t.Run("owner revokes", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{}
		target := uuid.New()
		if err := newShareSvc(permRepo, docRepo, nil, nil).RevokeAccess(docID, owner, target); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if permRepo.deletedUser != target {
			t.Fatal("expected the target permission to be deleted")
		}
	})

	t.Run("non owner is rejected", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(&fakePermRepo{}, docRepo, nil, nil).RevokeAccess(docID, uuid.New(), uuid.New()); err == nil {
			t.Fatal("expected a permission error")
		}
	})
}

func TestRequestAccess(t *testing.T) {
	owner := uuid.New()
	requester := uuid.New()
	docID := uuid.New()

	t.Run("missing doc", func(t *testing.T) {
		docRepo := &fakeDocRepo{byIDErr: gorm.ErrRecordNotFound}
		err := newShareSvc(&fakePermRepo{}, docRepo, nil, nil).RequestAccess(docID, requester, models.RoleViewer)
		if err == nil || err.Error() != "document not found" {
			t.Fatalf("expected document not found, got %v", err)
		}
	})

	t.Run("cannot request own doc", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		err := newShareSvc(&fakePermRepo{}, docRepo, nil, nil).RequestAccess(docID, owner, models.RoleViewer)
		if err == nil {
			t.Fatal("expected an error when requesting your own doc")
		}
	})

	t.Run("already has equal or higher access", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleEditor}}
		err := newShareSvc(permRepo, docRepo, nil, nil).RequestAccess(docID, requester, models.RoleViewer)
		if err == nil || err.Error() != "already has access" {
			t.Fatalf("expected already has access, got %v", err)
		}
	})

	t.Run("duplicate pending request is blocked", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{
			permErr:  gorm.ErrRecordNotFound,
			requests: []models.AccessRequest{{DocumentID: docID, Status: models.StatusPending, Role: models.RoleViewer}},
		}
		err := newShareSvc(permRepo, docRepo, nil, nil).RequestAccess(docID, requester, models.RoleViewer)
		if err == nil || err.Error() != "request already pending" {
			t.Fatalf("expected request already pending, got %v", err)
		}
	})

	t.Run("creates a pending request", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{permErr: gorm.ErrRecordNotFound, reqErr: nil}
		if err := newShareSvc(permRepo, docRepo, nil, nil).RequestAccess(docID, requester, models.RoleViewer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if permRepo.createdReq == nil || permRepo.createdReq.Status != models.StatusPending {
			t.Fatal("expected a pending request to be created")
		}
	})
}

func TestHandleAccessRequest(t *testing.T) {
	owner := uuid.New()
	requester := uuid.New()
	docID := uuid.New()

	t.Run("request not found", func(t *testing.T) {
		permRepo := &fakePermRepo{getReqErr: gorm.ErrRecordNotFound}
		if err := newShareSvc(permRepo, &fakeDocRepo{}, nil, nil).HandleAccessRequest(uuid.New(), owner, models.StatusApproved, models.RoleViewer); err == nil {
			t.Fatal("expected an error")
		}
	})

	t.Run("only the owner can handle", func(t *testing.T) {
		permRepo := &fakePermRepo{getReq: &models.AccessRequest{ID: uuid.New(), DocumentID: docID, Status: models.StatusPending}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(permRepo, docRepo, nil, nil).HandleAccessRequest(uuid.New(), uuid.New(), models.StatusApproved, models.RoleViewer); err == nil {
			t.Fatal("expected a permission error")
		}
	})

	t.Run("already handled", func(t *testing.T) {
		permRepo := &fakePermRepo{getReq: &models.AccessRequest{ID: uuid.New(), DocumentID: docID, Status: models.StatusApproved}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(permRepo, docRepo, nil, nil).HandleAccessRequest(uuid.New(), owner, models.StatusApproved, models.RoleViewer); err == nil {
			t.Fatal("expected an already handled error")
		}
	})

	t.Run("approve creates a permission", func(t *testing.T) {
		permRepo := &fakePermRepo{
			getReq:  &models.AccessRequest{ID: uuid.New(), DocumentID: docID, RequesterID: requester, Status: models.StatusPending},
			permErr: gorm.ErrRecordNotFound,
		}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(permRepo, docRepo, nil, nil).HandleAccessRequest(uuid.New(), owner, models.StatusApproved, models.RoleEditor); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if permRepo.createdPerm == nil || permRepo.createdPerm.Role != models.RoleEditor {
			t.Fatal("expected an editor permission to be created")
		}
	})

	t.Run("approve upgrades an existing permission", func(t *testing.T) {
		existing := &models.DocumentPermission{ID: uuid.New(), Role: models.RoleViewer}
		permRepo := &fakePermRepo{
			getReq: &models.AccessRequest{ID: uuid.New(), DocumentID: docID, RequesterID: requester, Status: models.StatusPending},
			perm:   existing,
		}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(permRepo, docRepo, nil, nil).HandleAccessRequest(uuid.New(), owner, models.StatusApproved, models.RoleEditor); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(permRepo.permUpdates) != 1 || permRepo.permUpdates[0].Role != models.RoleEditor {
			t.Fatal("expected the existing permission to be upgraded to editor")
		}
		if permRepo.createdPerm != nil {
			t.Fatal("should not create a second permission row")
		}
	})

	t.Run("reject does not touch permissions", func(t *testing.T) {
		permRepo := &fakePermRepo{getReq: &models.AccessRequest{ID: uuid.New(), DocumentID: docID, RequesterID: requester, Status: models.StatusPending}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newShareSvc(permRepo, docRepo, nil, nil).HandleAccessRequest(uuid.New(), owner, models.StatusRejected, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if permRepo.createdPerm != nil || len(permRepo.permUpdates) != 0 {
			t.Fatal("reject should not change permissions")
		}
	})
}

func TestListRequests(t *testing.T) {
	user := uuid.New()

	t.Run("returns sent and received", func(t *testing.T) {
		permRepo := &fakePermRepo{
			requests: []models.AccessRequest{{DocumentID: uuid.New()}},
			received: []models.AccessRequest{{DocumentID: uuid.New()}},
		}
		sent, received, err := newShareSvc(permRepo, &fakeDocRepo{}, nil, nil).ListRequests(user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sent) != 1 || len(received) != 1 {
			t.Fatalf("expected 1 sent and 1 received, got %d and %d", len(sent), len(received))
		}
	})

	t.Run("sent error", func(t *testing.T) {
		boom := errors.New("boom")
		permRepo := &fakePermRepo{reqErr: boom}
		if _, _, err := newShareSvc(permRepo, &fakeDocRepo{}, nil, nil).ListRequests(user); !errors.Is(err, boom) {
			t.Fatalf("expected error, got %v", err)
		}
	})

	t.Run("received error", func(t *testing.T) {
		boom := errors.New("boom")
		permRepo := &fakePermRepo{receivedErr: boom}
		if _, _, err := newShareSvc(permRepo, &fakeDocRepo{}, nil, nil).ListRequests(user); !errors.Is(err, boom) {
			t.Fatalf("expected error, got %v", err)
		}
	})
}

func TestAccessByToken(t *testing.T) {
	owner := uuid.New()
	accessor := uuid.New()
	docID := uuid.New()

	t.Run("invalid token", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byTokenErr: gorm.ErrRecordNotFound}
		_, err := newShareSvc(&fakePermRepo{}, &fakeDocRepo{}, nil, shareRepo).AccessByToken("bad", accessor)
		if err == nil || err.Error() != "invalid or expired token" {
			t.Fatalf("expected invalid token error, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byToken: &models.DocumentShare{DocumentID: docID, ExpiresAt: time.Now().Add(-time.Hour)}}
		_, err := newShareSvc(&fakePermRepo{}, &fakeDocRepo{}, nil, shareRepo).AccessByToken("t", accessor)
		if err == nil || err.Error() != "token expired" {
			t.Fatalf("expected token expired error, got %v", err)
		}
	})

	t.Run("missing document", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byToken: &models.DocumentShare{DocumentID: docID, ExpiresAt: time.Now().Add(time.Hour)}}
		docRepo := &fakeDocRepo{byIDErr: gorm.ErrRecordNotFound}
		_, err := newShareSvc(&fakePermRepo{}, docRepo, nil, shareRepo).AccessByToken("t", accessor)
		if err == nil || err.Error() != "document not found" {
			t.Fatalf("expected document not found, got %v", err)
		}
	})

	t.Run("owner gets the doc", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byToken: &models.DocumentShare{DocumentID: docID, ExpiresAt: time.Now().Add(time.Hour)}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		doc, err := newShareSvc(&fakePermRepo{}, docRepo, nil, shareRepo).AccessByToken("t", owner)
		if err != nil || doc.ID != docID {
			t.Fatalf("expected the doc for the owner, got doc=%v err=%v", doc, err)
		}
	})

	t.Run("user with a permission gets the doc", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byToken: &models.DocumentShare{DocumentID: docID, ExpiresAt: time.Now().Add(time.Hour)}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleViewer}}
		doc, err := newShareSvc(permRepo, docRepo, nil, shareRepo).AccessByToken("t", accessor)
		if err != nil || doc.ID != docID {
			t.Fatalf("expected the doc for a permission holder, got doc=%v err=%v", doc, err)
		}
	})

	t.Run("stranger is denied", func(t *testing.T) {
		shareRepo := &fakeShareRepo{byToken: &models.DocumentShare{DocumentID: docID, ExpiresAt: time.Now().Add(time.Hour)}}
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		permRepo := &fakePermRepo{permErr: gorm.ErrRecordNotFound}
		_, err := newShareSvc(permRepo, docRepo, nil, shareRepo).AccessByToken("t", accessor)
		if err == nil || err.Error() != "access denied: please request access from the document owner" {
			t.Fatalf("expected access denied, got %v", err)
		}
	})
}

func TestRoleRank(t *testing.T) {
	if roleRank(models.RoleEditor) <= roleRank(models.RoleViewer) {
		t.Fatal("editor should rank above viewer")
	}
}
