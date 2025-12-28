package services

import (
	"errors"

	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"github.com/personal-stakshi/draftly/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	LoginOrRegister(email, password, ip string) (*models.User, string, bool, error)
}

type authService struct {
	userRepo dbops.UserRepository
	config   *config.Config
}

func NewAuthService(repo dbops.UserRepository) AuthService {
	return &authService{
		userRepo: repo,
		config:   config.LoadConfig(),
	}
}

func (s *authService) LoginOrRegister(email, password, ip string) (*models.User, string, bool, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	isNewUser := false

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Registration Flow
		
		// 1. IP Rate Limiting Check
		ipUsage, err := s.userRepo.GetIPUsage(ip)
		if err == nil && ipUsage.CreatedAccounts >= 2 {
			return nil, "", false, errors.New("registration limit exceeded for this IP")
		}

		// 2. Hash Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", false, errors.New("failed to hash password")
		}

		user = &models.User{
			Email:        email,
			PasswordHash: string(hashedPassword),
			IPAddress:    ip,
		}

		if err := s.userRepo.CreateUser(user); err != nil {
			return nil, "", false, errors.New("failed to create user")
		}

		// 3. Update IP Usage
		s.userRepo.IncrementIPUsage(ip)
		isNewUser = true

	} else if err == nil {
		// Login Flow
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, "", false, errors.New("invalid credentials")
		}
	} else {
		return nil, "", false, err
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, s.config)
	if err != nil {
		return nil, "", false, errors.New("failed to generate token")
	}

	return user, token, isNewUser, nil
}
