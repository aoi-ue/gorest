package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"
	"github.com/piLinux/GoREST/lib/middleware"
	"github.com/piLinux/GoREST/service"
)

var db = database.GetDB()

// LoginPayload ...
type LoginPayload struct {
	Login string `json:"login"`
	Pass  string `json:"pass"`
}

// Login ...
func Login(ctx *gin.Context) {
	var payload LoginPayload
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	u, err := service.GetUserByEmail(payload.Login)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if u.Password != model.HashPass(payload.Pass) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	jwtValue, err := middleware.GetJWT(u.ID, u.Email)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"jwt": jwtValue})
}
