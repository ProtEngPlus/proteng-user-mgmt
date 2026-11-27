package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/models"
)

const (
	maxAuthFailures  = 5
	authLockDuration = 15 * time.Minute
)

type AuthAttemptRepository interface {
	Wait(key string, now time.Time) (time.Duration, error)
	RecordFailure(key string, now time.Time) error
	Reset(key string) error
	EnsureIndexes() error
}

type authAttemptRepository struct {
	collection *mongo.Collection
}

func NewAuthAttemptRepository() AuthAttemptRepository {
	return &authAttemptRepository{collection: database.GetCollection("auth_attempts")}
}

func authLockRemaining(lockedUntil, now time.Time) time.Duration {
	if remaining := lockedUntil.Sub(now); remaining > 0 {
		return remaining
	}
	return 0
}

func shouldLockAuth(attempts int) bool {
	return attempts >= maxAuthFailures
}

func (ar *authAttemptRepository) Wait(key string, now time.Time) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	var attempt models.AuthAttempt
	err := ar.collection.FindOne(ctx, bson.M{"key": key}).Decode(&attempt)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	if attempt.LockedUntil.IsZero() {
		return 0, nil
	}
	if remaining := authLockRemaining(attempt.LockedUntil, now); remaining > 0 {
		return remaining, nil
	}
	_, err = ar.collection.DeleteOne(ctx, bson.M{"key": key})
	return 0, err
}

func (ar *authAttemptRepository) RecordFailure(key string, now time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	expiresAt := now.Add(authLockDuration)
	update := bson.M{
		"$inc": bson.M{"attempts": 1},
		"$set": bson.M{"expires_at": expiresAt},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var attempt models.AuthAttempt
	if err := ar.collection.FindOneAndUpdate(ctx, bson.M{"key": key}, update, opts).Decode(&attempt); err != nil {
		return err
	}
	if !shouldLockAuth(attempt.Attempts) {
		return nil
	}

	_, err := ar.collection.UpdateOne(ctx, bson.M{"key": key}, bson.M{"$set": bson.M{"locked_until": expiresAt}})
	return err
}

func (ar *authAttemptRepository) Reset(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err := ar.collection.DeleteOne(ctx, bson.M{"key": key})
	return err
}

func (ar *authAttemptRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err := ar.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
	})
	if err != nil {
		return fmt.Errorf("create auth attempt indexes: %w", err)
	}
	return nil
}
