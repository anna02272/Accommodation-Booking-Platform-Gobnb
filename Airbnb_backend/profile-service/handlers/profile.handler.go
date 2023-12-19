package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
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

func (ph *ProfileHandler) CreateProfile(ctx *gin.Context) {
	var user *domain.User
	//
	//user.Name = html.EscapeString(user.Name)
	//user.Email = html.EscapeString(user.Email)
	//user.Username = html.EscapeString(user.Username)
	//user.Lastname = html.EscapeString(user.Lastname)
	//user.Address.Country = html.EscapeString(user.Address.Country)
	//user.Address.City = html.EscapeString(user.Address.City)
	//user.Address.Street = html.EscapeString(user.Address.Street)

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

func (ph *ProfileHandler) DeleteProfile(ctx *gin.Context) {
	email := ctx.Params.ByName("email")
	errP := ph.profileService.FindUserByEmail(email)
	if errP != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": errP.Error()})
		return
	}

	err := ph.profileService.DeleteUserProfile(email)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile deleted successfully"})
}
func (ph *ProfileHandler) UpdateUser(ctx *gin.Context) {
	var user *domain.User
	log.Println(user)
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	// Pozovi servis za unos korisnika
	err := ph.profileService.UpdateUser(user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile updated successfully"})
}
