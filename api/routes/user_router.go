package routes

import (
	"proteng-user-mgmt/api/controllers"
	"proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine, ur repositories.UserRepository) {
	uc := controllers.NewUserController(ur)

	router.GET("/users", uc.GetAllUsers)
	router.GET("/users/:id", uc.GetUser)
	router.POST("/users", uc.CreateUser)
	router.PUT("/users/:id", uc.UpdateUser)
	router.DELETE("/users/:id", uc.DeleteUser)
}
