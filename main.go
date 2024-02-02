package main

import (
	"os"
	"proteng-user-mgmt/apis/routes"
	"proteng-user-mgmt/configs"
	"proteng-user-mgmt/database"
	"proteng-user-mgmt/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	configs.AutomaticLoadEnv()

	router := gin.Default()

	// database
	err := database.ConnectToDB()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	userRepository := repositories.NewUserRepository()
	adminRepository := repositories.NewAdminRepository()

	//health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// routes
	routes.UserRoute(router, userRepository)
	routes.AuthRoute(router, userRepository)
	routes.AdminRoute(router, adminRepository)

	// start server
	httpPort := os.Getenv("HTTP_PORT")
	err = router.Run(":" + httpPort)
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
