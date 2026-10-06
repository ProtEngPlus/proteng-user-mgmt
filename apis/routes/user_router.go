package routes

import (
	"github.com/protengplus/proteng-user-mgmt/apis/controllers"
	"github.com/protengplus/proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.Engine, ur repositories.UserRepository) {
	uc := controllers.NewUserController(ur)

	router.GET("/users", uc.GetAllUsers)
	router.GET("/users/:id", uc.GetUser)
	router.PUT("/users/:id", uc.UpdateUser)
	router.DELETE("/users/:id", uc.DeleteUser)
}
