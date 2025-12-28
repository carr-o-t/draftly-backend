package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/services"
)

func LoginOrRegister(c *gin.Context) {
	// validator
	req, err := ValidateAuthRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// service
	repo := dbops.NewUserRepository()
	svc := services.NewAuthService(repo)
	
	user, token, isNew, err := svc.LoginOrRegister(req.Email, req.Password, c.ClientIP())

	// handle Response AND Error mapping
	SendAuthResponse(c, user, token, isNew, err)
}
