package main

import (
	"time"

	"github.com/protengplus/proteng-user-mgmt/apis/routes"
	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/internal/logger"
	"github.com/protengplus/proteng-user-mgmt/repositories"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.InitZap()
	configs.AutomaticLoadEnv()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// database
	err := database.ConnectToDB()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	userRepository := repositories.NewUserRepository()
	adminRepository := repositories.NewAdminRepository()

	// logging middleware
	router.Use(ginzap.GinzapWithConfig(logger.Zap, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/metrics", "/health"},
	}))

	//health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// routes
	routes.UserRoute(router, userRepository)
	routes.AuthRoute(router, userRepository, adminRepository)
	routes.AdminRoute(router, adminRepository)

	// panic recovery
	router.Use(ginzap.RecoveryWithZap(logger.Zap, true))

	// start server
	httpPort := configs.Config.HttpPort
	err = router.Run(":" + httpPort)
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
