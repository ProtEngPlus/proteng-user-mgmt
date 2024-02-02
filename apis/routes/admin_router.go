package routes

import (
	"proteng-user-mgmt/apis/controllers"
	"proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func AdminRoute(router *gin.Engine, ur repositories.AdminRepository) {
	uc := controllers.NewAdminController(ur)

	router.GET("/admins", uc.GetAllAdmins)
	router.GET("/admins/:id", uc.GetAdmin)
	router.POST("/admins", uc.CreateAdmin)
	router.PUT("/admins/:id", uc.UpdateAdmin)
	router.DELETE("/admins/:id", uc.DeleteAdmin)
}
