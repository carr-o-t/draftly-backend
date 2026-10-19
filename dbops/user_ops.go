package dbops

import (
	"errors"
	"time"

	"github.com/personal-stakshi/draftly/backend/database"
	"github.com/personal-stakshi/draftly/backend/models"
	"gorm.io/gorm"
)

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	return database.DB.Create(user).Error
}

func GetIPUsage(ip string) (*models.IPUsage, error) {
	var ipUsage models.IPUsage
	result := database.DB.Where("ip_address = ?", ip).First(&ipUsage)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ipUsage, nil
}

func CreatedIPVage(ipUsage *models.IPUsage) error {
	return database.DB.Create(ipUsage).Error
}

func UpdateIPUsage(ipUsage *models.IPUsage) error {
	return database.DB.Save(ipUsage).Error
}

func IncrementIPUsage(ip string) error {
	ipUsage, err := GetIPUsage(ip)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUsage := models.IPUsage{
			IPAddress:       ip,
			CreatedAccounts: 1,
			LastAttemptAt:   time.Now(),
		}
		return CreatedIPVage(&newUsage)
	} else if err != nil {
		return err
	}

	ipUsage.CreatedAccounts++
	ipUsage.LastAttemptAt = time.Now()
	return UpdateIPUsage(ipUsage)
}
