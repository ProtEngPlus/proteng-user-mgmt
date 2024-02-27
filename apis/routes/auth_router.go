package routes

import (
	"github.com/protengplus/proteng-user-mgmt/apis/controllers"
	"github.com/protengplus/proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, ur repositories.UserRepository, ar repositories.AdminRepository) {
	ac := controllers.NewAuthController(ur, ar)

	router.POST("/auth/login", ac.SignInUser)
	router.POST("/auth/login/admin", ac.SignInAdmin)
	router.POST("/auth/forgotpassword", ac.ForgotPassword)
	router.PATCH("/auth/resetpassword/:resetToken", ac.ResetPassword)

}
