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
		// new user, so sign them up
		
		// 1. make sure this ip has not signed up too many times
		ipUsage, err := s.userRepo.GetIPUsage(ip)
		if err == nil && ipUsage.CreatedAccounts >= 3 {
			return nil, "", false, errors.New("registration limit exceeded for this IP")
		}

		// 2. hash the password
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

		// 3. record the signup against this ip
		s.userRepo.IncrementIPUsage(ip)
		isNewUser = true

	} else if err == nil {
		// existing user, so log them in
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, "", false, errors.New("invalid credentials")
		}
	} else {
		return nil, "", false, err
	}

	// make a token
	token, err := utils.GenerateToken(user.ID, s.config)
	if err != nil {
		return nil, "", false, errors.New("failed to generate token")
	}

	return user, token, isNewUser, nil
}
