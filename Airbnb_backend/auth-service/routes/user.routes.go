package routes

import (
	"auth-service/handlers"
	"github.com/gin-gonic/gin"
)

type UserRouteHandler struct {
	userHandler handlers.UserHandler
}

func NewRouteUserHandler(userHandler handlers.UserHandler) UserRouteHandler {
	return UserRouteHandler{userHandler}
}

func (uc *UserRouteHandler) UserRoute(rg *gin.RouterGroup) {

	router := rg.Group("users")
	router.GET("/currentUser", uc.userHandler.CurrentUser)
	router.GET("/getById/:userId", uc.userHandler.GetUserById)

	router.PATCH("/changePassword", uc.userHandler.ChangePassword)
	router.DELETE("/delete", uc.userHandler.DeleteUser)

}
