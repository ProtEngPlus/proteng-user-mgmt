package routes

import (
	"proteng-user-mgmt/api/controllers"
	"proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, ur repositories.UserRepository) {
	ac := controllers.NewAuthController(ur)

	router.POST("/auth/login", ac.SignInUser)
}
