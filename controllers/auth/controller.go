package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-stakshi/draftly/backend/dbops"
	"github.com/personal-stakshi/draftly/backend/services"
)

func LoginOrRegister(c *gin.Context) {
	// check the input
	req, err := ValidateAuthRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// do the work
	repo := dbops.NewUserRepository()
	svc := services.NewAuthService(repo)
	
	user, token, isNew, err := svc.LoginOrRegister(req.Email, req.Password, c.ClientIP())

	// send back the result or the error
	SendAuthResponse(c, user, token, isNew, err)
}
