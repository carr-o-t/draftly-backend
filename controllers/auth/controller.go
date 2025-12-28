package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/config"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/models"
	"github.com/personal-stakshi/draftly/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
	IsNew bool        `json:"is_new"`
}

func LoginOrRegister(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	cfg := config.LoadConfig()
	clientIP := c.ClientIP()

	var user models.User
	result := dbops.DB.Where("email = ?", req.Email).First(&user)

	isNewUser := false

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Registration Flow
		
		// 1. IP Rate Limiting Check
		if err := checkIPLimit(clientIP); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "message": err.Error()})
			return
		}

		// 2. Create User
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to hash password"})
			return
		}

		user = models.User{
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			IPAddress:    clientIP,
		}

		if err := dbops.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create user"})
			return
		}

		// 3. Update IP Usage
		incrementIPUsage(clientIP)
		isNewUser = true

	} else if result.Error == nil {
		// Login Flow
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Database error"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true, 
		"data": AuthResponse{
			Token: token,
			User:  user,
			IsNew: isNewUser,
		},
	})
}

func checkIPLimit(ip string) error {
	var ipUsage models.IPUsage
	result := dbops.DB.Where("ip_address = ?", ip).First(&ipUsage)

	if result.Error == nil {
		if ipUsage.CreatedAccounts >= 2 {
			return errors.New("registration limit exceeded for this IP")
		}
	}
	return nil
}

func incrementIPUsage(ip string) {
	var ipUsage models.IPUsage
	result := dbops.DB.Where("ip_address = ?", ip).First(&ipUsage)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		ipUsage = models.IPUsage{
			IPAddress:       ip,
			CreatedAccounts: 1,
			LastAttemptAt:   time.Now(),
		}
		dbops.DB.Create(&ipUsage)
	} else {
		ipUsage.CreatedAccounts++
		ipUsage.LastAttemptAt = time.Now()
		dbops.DB.Save(&ipUsage)
	}
}
