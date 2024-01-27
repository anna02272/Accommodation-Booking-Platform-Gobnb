package handlers

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"profile-service/domain"
	"profile-service/services"
)

type ProfileHandler struct {
	profileService services.ProfileService
	Tracer         trace.Tracer
}

func NewProfileHandler(profileService services.ProfileService, tr trace.Tracer) ProfileHandler {
	return ProfileHandler{profileService, tr}
}

func (ph *ProfileHandler) CreateProfile(ctx *gin.Context) {
	spanCtx, span := ph.Tracer.Start(ctx.Request.Context(), "ProfileHandler.CreateProfile")
	defer span.End()

	var user *domain.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	err := ph.profileService.Registration(user, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Profile created successfully")
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile created successfully"})
}

func (ph *ProfileHandler) DeleteProfile(ctx *gin.Context) {
	spanCtx, span := ph.Tracer.Start(ctx.Request.Context(), "ProfileHandler.DeleteProfile")
	defer span.End()

	email := ctx.Params.ByName("email")
	errP := ph.profileService.FindUserByEmail(email, spanCtx)
	if errP != nil {
		span.SetStatus(codes.Error, errP.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": errP.Error()})
		return
	}

	err := ph.profileService.DeleteUserProfile(email, spanCtx)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Profile deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile deleted successfully"})
}
func (ph *ProfileHandler) UpdateUser(ctx *gin.Context) {
	spanCtx, span := ph.Tracer.Start(ctx.Request.Context(), "ProfileHandler.UpdateUser")
	defer span.End()

	var user *domain.User
	log.Println(user)
	if err := ctx.ShouldBindJSON(&user); err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	// Pozovi servis za unos korisnika
	err := ph.profileService.UpdateUser(user, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	span.SetStatus(codes.Ok, "Profile updated successfully")
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile updated successfully"})
}
func (ph *ProfileHandler) FindUserByEmail(ctx *gin.Context) {
	spanCtx, span := ph.Tracer.Start(ctx.Request.Context(), "ProfileHandler.FindUserByEmail")
	defer span.End()

	email := ctx.Param("email")

	if email == "" {
		span.SetStatus(codes.Error, "Email is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	user, err := ph.profileService.FindProfileByEmail(email, spanCtx)
	if err != nil {
		span.SetStatus(codes.Error, "User not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	span.SetStatus(codes.Ok, "Found user by email successfully")
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func ExtractTraceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
