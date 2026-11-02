package routes

import (
	"html/template"

	"github.com/protengplus/proteng-user-mgmt/apis/controllers"
	"github.com/protengplus/proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, ur repositories.UserRepository, ar repositories.AdminRepository, rr repositories.RefreshTokenRepository, temp *template.Template) {
	ac := controllers.NewAuthController(ur, ar, rr, temp)

	router.POST("/auth/register", ac.RegisterUser)
	router.POST("/auth/login", ac.SignInUser)
	router.POST("/auth/login/admin", ac.SignInAdmin)
	router.POST("/auth/logout", ac.Logout)
	router.POST("/auth/forgotpassword", ac.ForgotPassword)
	router.PATCH("/auth/resetpassword/:resetToken", ac.ResetPassword)
	router.PATCH("/auth/changepassword/:id", ac.ChangePassword)
	router.POST("/auth/sentverification", ac.SendVerification)
	router.POST("/auth/verifyemail/:verificationToken", ac.VerifyEmail)
}
