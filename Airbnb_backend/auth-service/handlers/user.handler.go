package handlers

import (
	"auth-service/domain"
	error2 "auth-service/error"
	"auth-service/services"
	"auth-service/utils"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"html"
	"net/http"
	"time"
)

type UserHandler struct {
	userService    services.UserService
	Tracer         trace.Tracer
	CircuitBreaker *gobreaker.CircuitBreaker
	logger         *logger.Logger
}

func NewUserHandler(userService services.UserService, tr trace.Tracer, logger *logger.Logger) UserHandler {
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "HTTPSRequest",
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker state changed from %s to %s\n", from, to)
		},
	})
	return UserHandler{
		userService:    userService,
		Tracer:         tr,
		CircuitBreaker: circuitBreaker,
		logger:         logger,
	}
}

var currentProfileUser *domain.User

func (ac *UserHandler) CurrentUserProfile(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentUserProfile")
	defer span.End()
	ac.logger.WithFields(logger.Fields{"path": "auth/currentUserProfile"}).Info("UserHandler.CurrentUserProfile")

	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ac.logger.WithFields(logger.Fields{"path": "auth/currentUserProfile"}).Error("Missing authorization header")
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)
	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/currentUserProfile"}).Error("Invalid token")
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}
	_, err = ac.userService.FindProfileInfoByEmail(spanCtx, user.Email)
	println(currentProfileUser)
	ac.logger.WithFields(logger.Fields{"path": "auth/currentUserProfile"}).Info("Current user retrieved successfully")
	span.SetStatus(codes.Ok, "Current user retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": currentProfileUser})
}

func (ac *UserHandler) CurrentUser(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentUser")
	defer span.End()
	ac.logger.WithFields(logger.Fields{"path": "auth/currentUser"}).Info("UserHandler.CurrentUser")

	tokenString := ctx.GetHeader("Authorization")
	tokenString = html.EscapeString(tokenString)

	if tokenString == "" {
		ac.logger.WithFields(logger.Fields{"path": "auth/currentUser"}).Error("Missing authorization header")
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)
	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/currentUser"}).Error("Invalid token")
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}
	ac.logger.WithFields(logger.Fields{"path": "auth/currentUser"}).Info("Current user retrieved successfully")
	span.SetStatus(codes.Ok, "Current user retrieved successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Token is valid", "user": user})
}

func GetUserFromToken(tokenString string, userService services.UserService, ctx context.Context, tracer trace.Tracer) (*domain.User, error) {
	_, span := tracer.Start(ctx, "UserHandler.CurrentUserProfile")
	defer span.End()

	tokenString = html.EscapeString(tokenString)

	if err := utils.VerifyToken(tokenString); err != nil {
		return nil, err
	}

	claims, err := utils.ParseTokenClaims(tokenString)
	if err != nil {
		return nil, err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("invalid username in token")
	}

	user, err := userService.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uh *UserHandler) GetUserById(ctx *gin.Context) {
	spanCtx, span := uh.Tracer.Start(ctx.Request.Context(), "UserHandler.GetUserById")
	defer span.End()
	uh.logger.WithFields(logger.Fields{"path": "auth/getUserById"}).Info("UserHandler.GetUserById")
	userID := ctx.Param("userId")

	if userID == "" {
		uh.logger.WithFields(logger.Fields{"path": "auth/getUserById"}).Error("User ID is required")
		span.SetStatus(codes.Error, "User ID is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	user, err := uh.userService.FindUserById(userID, spanCtx)
	if err != nil {
		uh.logger.WithFields(logger.Fields{"path": "auth/getUserById"}).Error("User not found")
		span.SetStatus(codes.Error, "User not found")
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	uh.logger.WithFields(logger.Fields{"path": "auth/getUserById"}).Info("Get user by id successful")
	span.SetStatus(codes.Ok, "Get user by id successful")
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (ac *UserHandler) ChangePassword(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.ChangePassword")
	ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Info("UserHandler.ChangePassword")
	defer span.End()

	var updatePassword *domain.PasswordChangeRequest

	tokenString := ctx.GetHeader("Authorization")

	if tokenString == "" {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Missing authorization header")

		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]

	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)

	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Invalid token")
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if err := ctx.ShouldBindJSON(&updatePassword); err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Invalid request body")
		span.SetStatus(codes.Error, "Invalid request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	passwordExistsBlackList, err := utils.CheckBlackList(updatePassword.NewPassword, "blacklist.txt")

	if passwordExistsBlackList {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Password is in blacklist")
		span.SetStatus(codes.Error, "Password is in blacklist!")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password is in blacklist!"})
		return
	}

	if !utils.ValidatePassword(updatePassword.NewPassword) {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Invalid password format")
		span.SetStatus(codes.Error, "Invalid password format")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid password format"})
		return
	}

	if updatePassword.NewPassword != updatePassword.ConfirmNewPassword {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Passwords do not match")
		span.SetStatus(codes.Error, "Passwords do not match")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}
	if err := utils.VerifyPassword(user.Password, updatePassword.CurrentPassword); err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Current password and new password do not match")
		span.SetStatus(codes.Error, "Current password and new password do not match")
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Current password and new password do not match"})
		return
	}
	hashedNewPassword, err := utils.HashPassword(updatePassword.NewPassword)
	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Failed to hash new password")
		span.SetStatus(codes.Error, "Failed to hash new password")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash new password"})
		return
	}

	user.Password = hashedNewPassword

	if err := ac.userService.UpdateUser(user, spanCtx); err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Error("Failed to update password")
		span.SetStatus(codes.Error, "Failed to update password")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}
	ac.logger.WithFields(logger.Fields{"path": "auth/changePassword"}).Info("Password updated saccussfully")
	span.SetStatus(codes.Ok, "Password updated successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (ac *UserHandler) DeleteUser(ctx *gin.Context) {
	spanCtx, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.DeleteUser")
	defer span.End()
	ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Info("UserHandler.DeleteUser")
	rw := ctx.Writer
	//h := ctx.Request

	tokenStringHeader := ctx.GetHeader("Authorization")
	tokenString := html.EscapeString(tokenStringHeader)

	if tokenString == "" {
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Missing authorization header")
		span.SetStatus(codes.Error, "Missing authorization header")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing authorization header"})
		return
	}
	tokenString = tokenString[len("Bearer "):]
	user, err := GetUserFromToken(tokenString, ac.userService, spanCtx, ac.Tracer)

	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Invalid token")
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	if user == nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Invalid token")
		span.SetStatus(codes.Error, "Invalid token")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		return
	}

	fmt.Println(user.UserRole)
	if user.UserRole == "Guest" {
		fmt.Println("here")

		urlCheckReservations := "https://res-server:8082/api/reservations/getAll"
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(spanCtx, timeout)
		defer cancel()

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				// Circuit is open
				ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Circuit is open. Reservation service is not available.")
				span.SetStatus(codes.Error, "Circuit is open. Reservation service is not available.")
				error2.ReturnJSONError(rw, "Reservation service is not available.", http.StatusBadRequest)
				return
			}

			if ctxRest.Err() == context.DeadlineExceeded {
				ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to fetch user reservations")
				span.SetStatus(codes.Error, "Failed to fetch user reservations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
				return
			}
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to fetch user reservations")
			span.SetStatus(codes.Error, "Failed to fetch user reservations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch user reservations"})
			return
		}
		defer respRes.Body.Close()

		if respRes.StatusCode != 404 {
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("You cannot delete your profile, you have active reservations")
			span.SetStatus(codes.Error, "You cannot delete your profile, you have active reservations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have active reservations"})
			return
		}
	}

	if user.UserRole == "Host" {
		userIDString := user.ID.Hex()
		fmt.Println(userIDString)
		urlCheckReservations := "https://acc-server:8083/api/accommodations/get/host/" + userIDString
		fmt.Println(urlCheckReservations)

		timeout := 2000 * time.Second // Adjust the timeout duration as needed
		ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		respRes, errRes := ac.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, tokenStringHeader, urlCheckReservations, "GET")
		if errRes != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Circuit is open. Accommodation service is not available.")
				span.SetStatus(codes.Error, "Circuit is open. Accommodation service is not available.")
				error2.ReturnJSONError(rw, "Accommodation service is not available.", http.StatusBadRequest)
				return
			}

			if ctxRest.Err() == context.DeadlineExceeded {
				ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to fetch host accommodations")
				span.SetStatus(codes.Error, "Failed to fetch host accommodations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
				return
			}
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to fetch host accommodations")
			span.SetStatus(codes.Error, "Failed to fetch host accommodations")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to fetch host accommodations"})
			return
		}
		defer respRes.Body.Close()

		var response map[string]interface{}
		if err := json.NewDecoder(respRes.Body).Decode(&response); err != nil {
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to decode response")
			span.SetStatus(codes.Error, "Failed to decode response")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to decode response"})
			return
		}

		if accommodations, ok := response["accommodations"].([]interface{}); ok {
			if len(accommodations) > 0 {
				ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("You cannot delete your profile, you have created accommodations")
				span.SetStatus(codes.Error, "You cannot delete your profile, you have created accommodations")
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "You cannot delete your profile, you have created accommodations"})
				return
			}
		}
	}
	urlProfile := "https://profile-server:8084/api/profile/delete/" + user.Email
	fmt.Println(urlProfile)

	timeout := 2000 * time.Second // Adjust the timeout duration as needed
	ctxRest, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, errRes := ac.HTTPSperformAuthorizationRequestWithCircuitBreaker(spanCtx, tokenStringHeader, urlProfile, "DELETE")
	if errRes != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			// Circuit is open
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Circuit is open. Reservation service is not available.")
			span.SetStatus(codes.Error, "Circuit is open. Reservation service is not available.")
			error2.ReturnJSONError(rw, "Reservation service is not available.", http.StatusBadRequest)
			return
		}

		if ctxRest.Err() == context.DeadlineExceeded {
			ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to delete user credentials")
			span.SetStatus(codes.Error, "Failed to delete user credentials")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
			return
		}
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to delete user credentials")
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to delete user credentials")
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}

	err = ac.userService.DeleteCredentials(user, spanCtx)
	if err != nil {
		ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Error("Failed to delete user credentials")
		span.SetStatus(codes.Error, "Failed to delete user credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to delete user credentials"})
		return
	}
	ac.logger.WithFields(logger.Fields{"path": "auth/deleteUser"}).Info("User deleted successfully")
	span.SetStatus(codes.Ok, "User deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (ac *UserHandler) HTTPSperformAuthorizationRequestWithContext(ctx context.Context, token string, url string, method string) (*http.Response, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	// Perform the request with the provided context
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (ac *UserHandler) HTTPSperformAuthorizationRequestWithCircuitBreaker(ctx context.Context, token string, url string, method string) (*http.Response, error) {
	maxRetries := 3

	retryOperation := func() (interface{}, error) {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", token)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		client := &http.Client{Transport: tr}
		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		return resp, nil // Return the response as the first value
	}

	result, err := ac.CircuitBreaker.Execute(func() (interface{}, error) {
		return retryOperationWithExponentialBackoff(ctx, maxRetries, retryOperation)
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("result here")
	fmt.Println(result)
	resp, ok := result.(*http.Response)
	if !ok {
		fmt.Println(ok)
		fmt.Println("OK")
		return nil, errors.New("unexpected response type from Circuit Breaker")
	}
	return resp, nil
}

func (ac *UserHandler) CurrentProfile(ctx *gin.Context) {
	_, span := ac.Tracer.Start(ctx.Request.Context(), "UserHandler.CurrentProfile")
	defer span.End()
	ac.logger.Info("UserHandler.CurrentProfile")
	var user *domain.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ac.logger.Errorf("Error: %v", err.Error())
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	currentProfileUser = user

}

func retryOperationWithExponentialBackoff(ctx context.Context, maxRetries int, operation func() (interface{}, error)) (interface{}, error) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Println("attempt loop: ")
		fmt.Println(attempt)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := operation()
		fmt.Println(result)
		if err == nil {
			fmt.Println("out of loop here")
			return result, nil
		}
		fmt.Printf("Attempt %d failed: %s\n", attempt, err.Error())
		backoff := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("max retries exceeded")
}
