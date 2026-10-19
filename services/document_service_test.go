package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

// fake repos
type fakeDocRepo struct {
	dbops.DocumentRepository

	count         int64
	countErr      error
	createErr     error
	created       *models.Document
	byID          *models.Document
	byIDErr       error
	withShares    *models.Document
	withSharesErr error
	updateErr     error
	updated       *models.Document
	deletedID     uuid.UUID
	deleteErr     error
	listOwner     []models.Document
	listOwnerErr  error
}

func (f *fakeDocRepo) CountDocumentsByOwner(uuid.UUID) (int64, error) {
	return f.count, f.countErr
}

func (f *fakeDocRepo) CreateDocument(doc *models.Document) error {
	if f.createErr != nil {
		return f.createErr
	}
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	f.created = doc
	return nil
}

func (f *fakeDocRepo) GetDocumentWithShares(uuid.UUID) (*models.Document, error) {
	return f.withShares, f.withSharesErr
}

func (f *fakeDocRepo) GetDocumentByID(uuid.UUID) (*models.Document, error) {
	return f.byID, f.byIDErr
}

func (f *fakeDocRepo) UpdateDocument(doc *models.Document) error {
	f.updated = doc
	return f.updateErr
}

func (f *fakeDocRepo) DeleteDocument(id uuid.UUID) error {
	f.deletedID = id
	return f.deleteErr
}

func (f *fakeDocRepo) ListDocumentsByOwner(uuid.UUID) ([]models.Document, error) {
	return f.listOwner, f.listOwnerErr
}

type fakeShareRepo struct {
	dbops.ShareRepository
	created    *models.DocumentShare
	byToken    *models.DocumentShare
	byTokenErr error
}

func (f *fakeShareRepo) CreateShare(share *models.DocumentShare) error {
	f.created = share
	return nil
}

func (f *fakeShareRepo) GetShareByToken(string) (*models.DocumentShare, error) {
	return f.byToken, f.byTokenErr
}

type fakePermRepo struct {
	dbops.PermissionRepository
	perm         *models.DocumentPermission
	permErr      error
	shared       []models.DocumentPermission
	sharedErr    error
	permUpdates  []*models.DocumentPermission
	createdPerm  *models.DocumentPermission
	createErr    error
	deletedUser  uuid.UUID
	requests     []models.AccessRequest
	reqErr       error
	createdReq   *models.AccessRequest
	getReq       *models.AccessRequest
	getReqErr    error
	listPerms    []models.DocumentPermission
	listPermsErr error
	received     []models.AccessRequest
	receivedErr  error
}

func (f *fakePermRepo) GetPermission(uuid.UUID, uuid.UUID) (*models.DocumentPermission, error) {
	return f.perm, f.permErr
}

func (f *fakePermRepo) ListSharedDocuments(uuid.UUID) ([]models.DocumentPermission, error) {
	return f.shared, f.sharedErr
}

func (f *fakePermRepo) UpdatePermission(perm *models.DocumentPermission) error {
	f.permUpdates = append(f.permUpdates, perm)
	return nil
}

func (f *fakePermRepo) CreatePermission(perm *models.DocumentPermission) error {
	f.createdPerm = perm
	return f.createErr
}

func (f *fakePermRepo) DeletePermission(_, userID uuid.UUID) error {
	f.deletedUser = userID
	return nil
}

func (f *fakePermRepo) ListAccessRequests(uuid.UUID) ([]models.AccessRequest, error) {
	return f.requests, f.reqErr
}

func (f *fakePermRepo) CreateAccessRequest(req *models.AccessRequest) error {
	f.createdReq = req
	return nil
}

func (f *fakePermRepo) GetAccessRequest(uuid.UUID) (*models.AccessRequest, error) {
	return f.getReq, f.getReqErr
}

func (f *fakePermRepo) UpdateAccessRequest(*models.AccessRequest) error {
	return nil
}

func (f *fakePermRepo) ListPermissions(uuid.UUID) ([]models.DocumentPermission, error) {
	return f.listPerms, f.listPermsErr
}

func (f *fakePermRepo) ListPendingRequestsForDocs(uuid.UUID) ([]models.AccessRequest, error) {
	return f.received, f.receivedErr
}

func newDocService(d *fakeDocRepo, s *fakeShareRepo, p *fakePermRepo) DocumentService {
	return NewDocumentService(d, s, p)
}

func TestCreateDocument(t *testing.T) {
	owner := uuid.New()

	t.Run("creates with a default title and a share link", func(t *testing.T) {
		docRepo := &fakeDocRepo{}
		shareRepo := &fakeShareRepo{}

		doc, err := newDocService(docRepo, shareRepo, &fakePermRepo{}).CreateDocument(owner, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Title != "Untitled Document" {
			t.Fatalf("expected default title, got %q", doc.Title)
		}
		if doc.OwnerID != owner {
			t.Fatal("owner not set on the doc")
		}
		if shareRepo.created == nil || shareRepo.created.Token == "" {
			t.Fatal("expected a share link to be created")
		}
	})

	t.Run("fails when the user is at the doc limit", func(t *testing.T) {
		docRepo := &fakeDocRepo{count: MaxDocumentsPerUser}
		_, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).CreateDocument(owner, "x")
		if !errors.Is(err, ErrDocumentLimitReached) {
			t.Fatalf("expected limit error, got %v", err)
		}
	})

	t.Run("propagates the count error", func(t *testing.T) {
		boom := errors.New("boom")
		docRepo := &fakeDocRepo{countErr: boom}
		if _, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).CreateDocument(owner, "x"); !errors.Is(err, boom) {
			t.Fatalf("expected count error, got %v", err)
		}
	})

	t.Run("propagates the create error", func(t *testing.T) {
		boom := errors.New("boom")
		docRepo := &fakeDocRepo{createErr: boom}
		if _, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).CreateDocument(owner, "x"); !errors.Is(err, boom) {
			t.Fatalf("expected create error, got %v", err)
		}
	})
}

func TestGetDocumentRole(t *testing.T) {
	user := uuid.New()

	t.Run("owner", func(t *testing.T) {
		docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: user}}
		_, role, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).GetDocument(uuid.New(), user)
		if err != nil || role != "owner" {
			t.Fatalf("expected owner, got role=%q err=%v", role, err)
		}
	})

	t.Run("shared editor", func(t *testing.T) {
		docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: uuid.New()}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{Role: models.RoleEditor}}
		_, role, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).GetDocument(uuid.New(), user)
		if err != nil || role != "editor" {
			t.Fatalf("expected editor, got role=%q err=%v", role, err)
		}
	})

	t.Run("no access", func(t *testing.T) {
		docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: uuid.New()}}
		permRepo := &fakePermRepo{permErr: gorm.ErrRecordNotFound}
		_, _, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).GetDocument(uuid.New(), user)
		if err == nil || err.Error() != "access denied" {
			t.Fatalf("expected access denied, got %v", err)
		}
	})

	t.Run("missing document", func(t *testing.T) {
		docRepo := &fakeDocRepo{withSharesErr: gorm.ErrRecordNotFound}
		_, _, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).GetDocument(uuid.New(), user)
		if err == nil || err.Error() != "document not found" {
			t.Fatalf("expected document not found, got %v", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		boom := errors.New("boom")
		docRepo := &fakeDocRepo{withSharesErr: boom}
		if _, _, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).GetDocument(uuid.New(), user); !errors.Is(err, boom) {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}

func TestListDocuments(t *testing.T) {
	user := uuid.New()

	t.Run("splits owned and shared", func(t *testing.T) {
		docRepo := &fakeDocRepo{listOwner: []models.Document{{Title: "mine"}}}
		permRepo := &fakePermRepo{shared: []models.DocumentPermission{
			{Document: models.Document{Title: "theirs"}},
		}}
		mine, shared, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).ListDocuments(user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mine) != 1 || mine[0].Title != "mine" {
			t.Fatalf("owned docs wrong: %+v", mine)
		}
		if len(shared) != 1 || shared[0].Title != "theirs" {
			t.Fatalf("shared docs wrong: %+v", shared)
		}
	})

	t.Run("owner list error", func(t *testing.T) {
		boom := errors.New("boom")
		docRepo := &fakeDocRepo{listOwnerErr: boom}
		if _, _, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).ListDocuments(user); !errors.Is(err, boom) {
			t.Fatalf("expected error, got %v", err)
		}
	})

	t.Run("shared list error", func(t *testing.T) {
		boom := errors.New("boom")
		permRepo := &fakePermRepo{sharedErr: boom}
		if _, _, err := newDocService(&fakeDocRepo{}, &fakeShareRepo{}, permRepo).ListDocuments(user); !errors.Is(err, boom) {
			t.Fatalf("expected error, got %v", err)
		}
	})
}

func TestUpdateDocumentSanitizesContent(t *testing.T) {
	owner := uuid.New()
	docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: owner}}
	doc, err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).
		UpdateDocument(uuid.New(), owner, "New Title", `<p>hi</p><script>alert(1)</script>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "New Title" {
		t.Fatalf("title not updated: %q", doc.Title)
	}
	if strings.Contains(doc.Content, "<script") {
		t.Fatalf("content was not sanitized: %q", doc.Content)
	}
}

func TestUpdateDocumentPermissionDenied(t *testing.T) {
	viewer := uuid.New()
	docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: uuid.New()}}
	permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleViewer}}
	_, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).UpdateDocument(uuid.New(), viewer, "t", "c")
	if err == nil || err.Error() != "permission denied" {
		t.Fatalf("expected permission denied, got %v", err)
	}
}

func TestAutoSaveSanitizesAndChecksRole(t *testing.T) {
	editor := uuid.New()

	t.Run("editor can save", func(t *testing.T) {
		docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: uuid.New()}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleEditor}}
		doc, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).AutoSave(uuid.New(), editor, `<p>ok</p><iframe src="x"></iframe>`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(doc.Content, "<iframe") {
			t.Fatalf("autosave content was not sanitized: %q", doc.Content)
		}
	})

	t.Run("viewer cannot save", func(t *testing.T) {
		docRepo := &fakeDocRepo{withShares: &models.Document{OwnerID: uuid.New()}}
		permRepo := &fakePermRepo{perm: &models.DocumentPermission{ID: uuid.New(), Role: models.RoleViewer}}
		_, err := newDocService(docRepo, &fakeShareRepo{}, permRepo).AutoSave(uuid.New(), editor, "x")
		if err == nil || err.Error() != "permission denied" {
			t.Fatalf("expected permission denied, got %v", err)
		}
	})
}

func TestDeleteDocument(t *testing.T) {
	owner := uuid.New()
	docID := uuid.New()

	t.Run("owner can delete", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		if err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).DeleteDocument(docID, owner); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if docRepo.deletedID != docID {
			t.Fatalf("expected doc %s to be deleted", docID)
		}
	})

	t.Run("non owner is rejected", func(t *testing.T) {
		docRepo := &fakeDocRepo{byID: &models.Document{ID: docID, OwnerID: owner}}
		err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).DeleteDocument(docID, uuid.New())
		if err == nil {
			t.Fatal("expected a permission error")
		}
		if docRepo.deletedID != uuid.Nil {
			t.Fatal("doc should not have been deleted")
		}
	})

	t.Run("missing doc", func(t *testing.T) {
		docRepo := &fakeDocRepo{byIDErr: gorm.ErrRecordNotFound}
		if err := newDocService(docRepo, &fakeShareRepo{}, &fakePermRepo{}).DeleteDocument(docID, owner); err == nil {
			t.Fatal("expected an error for a missing doc")
		}
	})
}
