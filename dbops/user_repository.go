package dbops

import (
	"errors"
	"time"

	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	GetIPUsage(ip string) (*models.IPUsage, error)
	CreateIPUsage(ipUsage *models.IPUsage) error
	UpdateIPUsage(ipUsage *models.IPUsage) error
	IncrementIPUsage(ip string) error
}

type postgresUserRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &postgresUserRepository{db: database.DB}
}

func (r *postgresUserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *postgresUserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *postgresUserRepository) GetIPUsage(ip string) (*models.IPUsage, error) {
	var ipUsage models.IPUsage
	result := r.db.Where("ip_address = ?", ip).First(&ipUsage)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ipUsage, nil
}

func (r *postgresUserRepository) CreateIPUsage(ipUsage *models.IPUsage) error {
	return r.db.Create(ipUsage).Error
}

func (r *postgresUserRepository) UpdateIPUsage(ipUsage *models.IPUsage) error {
	return r.db.Save(ipUsage).Error
}

func (r *postgresUserRepository) IncrementIPUsage(ip string) error {
	ipUsage, err := r.GetIPUsage(ip)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUsage := models.IPUsage{
			IPAddress:       ip,
			CreatedAccounts: 1,
			LastAttemptAt:   time.Now(),
		}
		return r.CreateIPUsage(&newUsage)
	} else if err != nil {
		return err
	}

	ipUsage.CreatedAccounts++
	ipUsage.LastAttemptAt = time.Now()
	return r.UpdateIPUsage(ipUsage)
}
