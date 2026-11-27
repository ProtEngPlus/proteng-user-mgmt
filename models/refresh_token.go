package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	TokenHash string             `bson:"token_hash"`
	FamilyId  primitive.ObjectID `bson:"family_id"`
	SubjectId primitive.ObjectID `bson:"subject_id"`
	Role      string             `bson:"role"`
	ExpiresAt time.Time          `bson:"expires_at"`
	RevokedAt *time.Time         `bson:"revoked_at,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
}
