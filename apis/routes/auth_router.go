package routes

import (
	"proteng-user-mgmt/apis/controllers"
	"proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, ur repositories.UserRepository, ar repositories.AdminRepository) {
	ac := controllers.NewAuthController(ur, ar)

	router.POST("/auth/login", ac.SignInUser)
	router.POST("/auth/login/admin", ac.SignInAdmin)
}
