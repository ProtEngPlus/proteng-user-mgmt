package database

import (
	"context"
	"fmt"
	"time"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/internal/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const QueryTimeout = 5 * time.Second

var Client *mongo.Client

func ConnectToDB() error {
	uri := configs.Config.MongoUri
	logger.Zap.Info("Connecting to MongoDB")
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}

	Client = client
	logger.Zap.Info("Connected to MongoDB")
	return nil
}

func ConnectWithRetry(connect func() error, maxAttempts int, backoff time.Duration) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = connect(); err == nil {
			return nil
		}
		logger.Errorf("Failed to connect to MongoDB (attempt %d/%d): %v", attempt, maxAttempts, err)
		if attempt < maxAttempts {
			time.Sleep(backoff)
		}
	}
	return fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", maxAttempts, err)
}

func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database(configs.Config.MongoDb).Collection(collectionName)
}
