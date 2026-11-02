package repositories

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/internal/logger"
	"github.com/protengplus/proteng-user-mgmt/models"
)

func TestMain(m *testing.M) {
	logger.InitZap()
	os.Exit(m.Run())
}

func newTestRefreshTokenRepository(t *testing.T) RefreshTokenRepository {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("set MONGO_TEST_URI to run repository tests against a real MongoDB")
	}
	configs.Config.MongoUri = uri
	configs.Config.MongoDb = "proteng_repository_test"
	if err := database.ConnectToDB(); err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		database.GetCollection("refresh_tokens").Drop(context.Background())
	})

	repo := NewRefreshTokenRepository()
	if err := repo.EnsureIndexes(); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	return repo
}

func TestRefreshTokenRepository(t *testing.T) {
	repo := newTestRefreshTokenRepository(t)
	now := time.Now()

	create := func(t *testing.T, hash string, subject, family primitive.ObjectID, expiresAt time.Time) {
		t.Helper()
		err := repo.Create(&models.RefreshToken{
			TokenHash: hash,
			FamilyId:  family,
			SubjectId: subject,
			Role:      "user",
			ExpiresAt: expiresAt,
			CreatedAt: now,
		})
		if err != nil {
			t.Fatalf("create %s: %v", hash, err)
		}
	}

	t.Run("consume works once then reports reuse", func(t *testing.T) {
		family := primitive.NewObjectID()
		create(t, "hash-once", primitive.NewObjectID(), family, now.Add(time.Hour))

		token, err := repo.Consume("hash-once", now)
		if err != nil || token.FamilyId != family {
			t.Fatalf("first consume = %+v, %v", token, err)
		}
		token, err = repo.Consume("hash-once", now)
		if !errors.Is(err, ErrRefreshTokenReused) || token == nil || token.RevokedAt == nil {
			t.Fatalf("second consume = %+v, %v; want ErrRefreshTokenReused", token, err)
		}
	})

	t.Run("unknown token is invalid", func(t *testing.T) {
		if _, err := repo.Consume("hash-unknown", now); !errors.Is(err, ErrRefreshTokenInvalid) {
			t.Fatalf("err = %v, want ErrRefreshTokenInvalid", err)
		}
	})

	t.Run("expired token is invalid", func(t *testing.T) {
		create(t, "hash-expired", primitive.NewObjectID(), primitive.NewObjectID(), now.Add(-time.Minute))
		if _, err := repo.Consume("hash-expired", now); !errors.Is(err, ErrRefreshTokenInvalid) {
			t.Fatalf("err = %v, want ErrRefreshTokenInvalid", err)
		}
	})

	t.Run("revoke family revokes every token in it", func(t *testing.T) {
		family := primitive.NewObjectID()
		create(t, "hash-family-a", primitive.NewObjectID(), family, now.Add(time.Hour))
		create(t, "hash-family-b", primitive.NewObjectID(), family, now.Add(time.Hour))

		if err := repo.RevokeFamily(family, now); err != nil {
			t.Fatalf("revoke family: %v", err)
		}
		for _, hash := range []string{"hash-family-a", "hash-family-b"} {
			if _, err := repo.Consume(hash, now); !errors.Is(err, ErrRefreshTokenReused) {
				t.Errorf("%s: err = %v, want ErrRefreshTokenReused", hash, err)
			}
		}
	})

	t.Run("revoke all for subject leaves other subjects alone", func(t *testing.T) {
		subject := primitive.NewObjectID()
		create(t, "hash-subject-1", subject, primitive.NewObjectID(), now.Add(time.Hour))
		create(t, "hash-subject-2", subject, primitive.NewObjectID(), now.Add(time.Hour))
		create(t, "hash-other", primitive.NewObjectID(), primitive.NewObjectID(), now.Add(time.Hour))

		if err := repo.RevokeAllForSubject(subject, now); err != nil {
			t.Fatalf("revoke subject: %v", err)
		}
		for _, hash := range []string{"hash-subject-1", "hash-subject-2"} {
			if _, err := repo.Consume(hash, now); !errors.Is(err, ErrRefreshTokenReused) {
				t.Errorf("%s: err = %v, want ErrRefreshTokenReused", hash, err)
			}
		}
		if _, err := repo.Consume("hash-other", now); err != nil {
			t.Errorf("other subject's token should still work: %v", err)
		}
	})

	t.Run("find by hash", func(t *testing.T) {
		create(t, "hash-find", primitive.NewObjectID(), primitive.NewObjectID(), now.Add(time.Hour))
		if token, err := repo.FindByHash("hash-find"); err != nil || token.TokenHash != "hash-find" {
			t.Fatalf("find = %+v, %v", token, err)
		}
		if _, err := repo.FindByHash("hash-missing"); !errors.Is(err, ErrRefreshTokenInvalid) {
			t.Fatalf("missing err = %v, want ErrRefreshTokenInvalid", err)
		}
	})
}
