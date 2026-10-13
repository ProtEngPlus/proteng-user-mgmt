package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
)

type fakeRefreshTokenRepository struct {
	repositories.RefreshTokenRepository
	created   []*models.RefreshToken
	createErr error
}

func (f *fakeRefreshTokenRepository) Create(token *models.RefreshToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, token)
	return nil
}

func useTestSigningKey(t *testing.T) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	block := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	configs.Config.AccessTokenPrivateKey = base64.StdEncoding.EncodeToString(block)
	configs.Config.AccessTokenExpiredIn = "15m"
	configs.Config.RefreshTokenExpiredIn = "168h"
}

func TestIssueSession(t *testing.T) {
	useTestSigningKey(t)

	t.Run("stores only the hash of the refresh token", func(t *testing.T) {
		repo := &fakeRefreshTokenRepository{}
		ac := &AuthController{refreshTokenRepository: repo}
		subject, family := primitive.NewObjectID(), primitive.NewObjectID()

		session, err := ac.issueSession(subject, "user", family)
		if err != nil {
			t.Fatalf("issueSession() error = %v", err)
		}
		if session.AccessToken == "" || session.RefreshToken == "" {
			t.Fatalf("empty token in %+v", session)
		}
		if len(repo.created) != 1 {
			t.Fatalf("stored %d tokens, want 1", len(repo.created))
		}

		stored := repo.created[0]
		if stored.TokenHash != utils.HashRefreshToken(session.RefreshToken) || stored.TokenHash == session.RefreshToken {
			t.Errorf("stored hash %q does not match the raw token", stored.TokenHash)
		}
		if stored.FamilyId != family || stored.SubjectId != subject || stored.Role != "user" {
			t.Errorf("unexpected stored token %+v", stored)
		}
		if d := time.Until(stored.ExpiresAt); d < 167*time.Hour || d > 168*time.Hour {
			t.Errorf("refresh token expires in %v, want about 168h", d)
		}
	})

	t.Run("fails when the token cannot be stored", func(t *testing.T) {
		ac := &AuthController{refreshTokenRepository: &fakeRefreshTokenRepository{createErr: errors.New("db down")}}
		if _, err := ac.issueSession(primitive.NewObjectID(), "user", primitive.NewObjectID()); err == nil {
			t.Fatal("issueSession() error = nil, want error")
		}
	})

	t.Run("fails on a bad refresh ttl", func(t *testing.T) {
		configs.Config.RefreshTokenExpiredIn = "seven days"
		t.Cleanup(func() { configs.Config.RefreshTokenExpiredIn = "168h" })
		ac := &AuthController{refreshTokenRepository: &fakeRefreshTokenRepository{}}
		if _, err := ac.issueSession(primitive.NewObjectID(), "user", primitive.NewObjectID()); err == nil {
			t.Fatal("issueSession() error = nil, want error")
		}
	})
}
