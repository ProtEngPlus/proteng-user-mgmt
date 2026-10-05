package main

import (
	"html/template"
	"time"

	"github.com/protengplus/proteng-user-mgmt/apis/routes"
	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/internal/logger"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"go.uber.org/zap"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	rmqConsumer "github.com/protengplus/proteng-user-mgmt/internal/rabbitmq/consumer"
)

func main() {
	logger.InitZap()
	configs.AutomaticLoadEnv()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// database
	err := database.ConnectWithRetry(database.ConnectToDB, 5, 10*time.Second)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	userRepository := repositories.NewUserRepository()
	adminRepository := repositories.NewAdminRepository()

	if err := userRepository.EnsureIndexes(); err != nil {
		logger.Zap.Error("failed to ensure user email index", zap.Error(err))
	}
	if err := adminRepository.EnsureIndexes(); err != nil {
		logger.Zap.Error("failed to ensure admin email index", zap.Error(err))
	}

	temp := template.Must(template.ParseGlob("templates/*.html"))

	// rabbitmq
	rabbitConsumer := rmqConsumer.NewConsumer(userRepository, temp)

	amqpURL := configs.Config.RabbitMqUrl
	go func() {
		err := rabbitConsumer.RunConsumer(amqpURL, configs.Config.JobQueue)
		if err != nil {
			logger.Fatalf("Error in RabbitMQ Consumer: %v", err)
		}
	}()

	// logging middleware
	router.Use(ginzap.GinzapWithConfig(logger.Zap, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/metrics", "/health"},
	}))

	//health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok", "environment": configs.Config.Env})
	})

	// routes
	routes.UserRoute(router, userRepository)
	routes.AuthRoute(router, userRepository, adminRepository, temp)
	routes.AdminRoute(router, adminRepository)

	// panic recovery
	router.Use(ginzap.RecoveryWithZap(logger.Zap, true))

	// start server
	httpPort := configs.Config.HttpPort
	logger.Zap.Info("proteng-user-mgmt is running on :" + httpPort)
	err = router.Run(":" + httpPort)
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
