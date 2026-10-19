package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type fakeUserRepo struct {
	dbops.UserRepository

	user       *models.User
	userErr    error
	createErr  error
	created    *models.User
	ipUsage    *models.IPUsage
	ipUsageErr error
	increments int
}

func (f *fakeUserRepo) GetUserByEmail(string) (*models.User, error) {
	return f.user, f.userErr
}

func (f *fakeUserRepo) CreateUser(user *models.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	f.created = user
	return nil
}

func (f *fakeUserRepo) GetIPUsage(string) (*models.IPUsage, error) {
	return f.ipUsage, f.ipUsageErr
}

func (f *fakeUserRepo) IncrementIPUsage(string) error {
	f.increments++
	return nil
}

func newAuthService(repo *fakeUserRepo) AuthService {
	return &authService{userRepo: repo, config: &config.Config{JWTSecret: "test-secret"}}
}

func TestLoginOrRegisterNewUser(t *testing.T) {
	repo := &fakeUserRepo{userErr: gorm.ErrRecordNotFound, ipUsageErr: gorm.ErrRecordNotFound}

	user, token, isNew, err := newAuthService(repo).LoginOrRegister("new@example.com", "password123", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isNew {
		t.Fatal("expected isNew to be true")
	}
	if token == "" {
		t.Fatal("expected a token")
	}
	if user.PasswordHash == "password123" {
		t.Fatal("password was stored in plain text")
	}
	if repo.created == nil || repo.created.IPAddress != "1.2.3.4" {
		t.Fatal("ip address was not recorded on the user")
	}
	if repo.increments != 1 {
		t.Fatalf("expected ip usage to be incremented once, got %d", repo.increments)
	}
}

func TestLoginOrRegisterExistingUser(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &fakeUserRepo{user: &models.User{ID: uuid.New(), Email: "old@example.com", PasswordHash: string(hash)}}

	_, token, isNew, err := newAuthService(repo).LoginOrRegister("old@example.com", "password123", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isNew {
		t.Fatal("expected isNew to be false for an existing user")
	}
	if token == "" {
		t.Fatal("expected a token")
	}
}

func TestLoginOrRegisterWrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &fakeUserRepo{user: &models.User{ID: uuid.New(), Email: "old@example.com", PasswordHash: string(hash)}}

	_, _, _, err := newAuthService(repo).LoginOrRegister("old@example.com", "nope", "1.2.3.4")
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLoginOrRegisterHitRegistrationLimit(t *testing.T) {
	repo := &fakeUserRepo{
		userErr: gorm.ErrRecordNotFound,
		ipUsage: &models.IPUsage{IPAddress: "1.2.3.4", CreatedAccounts: 3},
	}

	_, _, _, err := newAuthService(repo).LoginOrRegister("new@example.com", "password123", "1.2.3.4")
	if err == nil || err.Error() != "registration limit exceeded for this IP" {
		t.Fatalf("expected the registration limit error, got %v", err)
	}
	if repo.increments != 0 {
		t.Fatal("should not increment usage when the limit is hit")
	}
}

func TestLoginOrRegisterRepoError(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeUserRepo{userErr: boom}

	if _, _, _, err := newAuthService(repo).LoginOrRegister("x@example.com", "password123", "ip"); !errors.Is(err, boom) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestLoginOrRegisterCreateUserFails(t *testing.T) {
	repo := &fakeUserRepo{userErr: gorm.ErrRecordNotFound, ipUsageErr: gorm.ErrRecordNotFound, createErr: errors.New("db down")}

	_, _, _, err := newAuthService(repo).LoginOrRegister("new@example.com", "password123", "ip")
	if err == nil || err.Error() != "failed to create user" {
		t.Fatalf("expected failed to create user, got %v", err)
	}
}
