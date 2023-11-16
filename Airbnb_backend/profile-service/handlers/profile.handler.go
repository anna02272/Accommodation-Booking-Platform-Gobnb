package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"profile-service/domain"
	"profile-service/services"
)

type ProfileHandler struct {
	profileService services.ProfileService
}

func NewProfileHandler(profileService services.ProfileService) ProfileHandler {
	return ProfileHandler{profileService}
}

// ... (prethodni kod)

func (ph *ProfileHandler) CreateProfile(ctx *gin.Context) {
	var user *domain.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	// Pozovi servis za unos korisnika
	err := ph.profileService.Registration(user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile created successfully"})
}
