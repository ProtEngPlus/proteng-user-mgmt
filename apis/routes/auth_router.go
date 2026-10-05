package routes

import (
	"html/template"

	"github.com/protengplus/proteng-user-mgmt/apis/controllers"
	"github.com/protengplus/proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, ur repositories.UserRepository, ar repositories.AdminRepository, temp *template.Template) {
	ac := controllers.NewAuthController(ur, ar, temp)

	router.POST("/auth/register", ac.RegisterUser)
	router.POST("/auth/login", ac.SignInUser)
	router.POST("/auth/login/admin", ac.SignInAdmin)
	router.POST("/auth/forgotpassword", ac.ForgotPassword)
	router.PATCH("/auth/resetpassword/:resetToken", ac.ResetPassword)
	router.PATCH("/auth/changepassword/:id", ac.ChangePassword)
	router.POST("/auth/sentverification", ac.SendVerification)
	router.POST("/auth/verifyemail/:verificationToken", ac.VerifyEmail)
}
