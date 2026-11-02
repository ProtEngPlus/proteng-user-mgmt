package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/models"
)

var (
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid or expired")
	ErrRefreshTokenReused  = errors.New("refresh token was already used")
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(hash string) (*models.RefreshToken, error)
	Consume(hash string, now time.Time) (*models.RefreshToken, error)
	RevokeFamily(familyId primitive.ObjectID, now time.Time) error
	RevokeAllForSubject(subjectId primitive.ObjectID, now time.Time) error
	EnsureIndexes() error
}

type refreshTokenRepository struct {
	collection *mongo.Collection
}

func NewRefreshTokenRepository() RefreshTokenRepository {
	return &refreshTokenRepository{collection: database.GetCollection("refresh_tokens")}
}

func (rr *refreshTokenRepository) Create(token *models.RefreshToken) error {
	token.Id = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	_, err := rr.collection.InsertOne(ctx, token)
	return err
}

func (rr *refreshTokenRepository) FindByHash(hash string) (*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	var token models.RefreshToken
	err := rr.collection.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&token)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrRefreshTokenInvalid
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (rr *refreshTokenRepository) Consume(hash string, now time.Time) (*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	filter := bson.M{
		"token_hash": hash,
		"revoked_at": bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": now},
	}
	update := bson.M{"$set": bson.M{"revoked_at": now}}

	var token models.RefreshToken
	err := rr.collection.FindOneAndUpdate(ctx, filter, update).Decode(&token)
	if err == nil {
		return &token, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	stored, err := rr.FindByHash(hash)
	if err != nil {
		return nil, err
	}
	if stored.RevokedAt != nil {
		return stored, ErrRefreshTokenReused
	}

	return nil, ErrRefreshTokenInvalid
}

func (rr *refreshTokenRepository) RevokeFamily(familyId primitive.ObjectID, now time.Time) error {
	return rr.revokeWhere(bson.M{"family_id": familyId}, now)
}

func (rr *refreshTokenRepository) RevokeAllForSubject(subjectId primitive.ObjectID, now time.Time) error {
	return rr.revokeWhere(bson.M{"subject_id": subjectId}, now)
}

func (rr *refreshTokenRepository) revokeWhere(filter bson.M, now time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	filter["revoked_at"] = bson.M{"$exists": false}
	_, err := rr.collection.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"revoked_at": now}})
	return err
}

func (rr *refreshTokenRepository) EnsureIndexes() error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "family_id", Value: 1}}},
		{Keys: bson.D{{Key: "subject_id", Value: 1}}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.QueryTimeout)
	defer cancel()

	if _, err := rr.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create refresh token indexes: %w", err)
	}

	return nil
}
