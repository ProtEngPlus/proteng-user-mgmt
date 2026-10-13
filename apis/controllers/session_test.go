package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/internal/logger"
	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
)

func TestMain(m *testing.M) {
	logger.InitZap()
	os.Exit(m.Run())
}

type fakeRefreshTokenRepository struct {
	repositories.RefreshTokenRepository
	created       []*models.RefreshToken
	createErr     error
	byHash        map[string]*models.RefreshToken
	revokedFamily primitive.ObjectID
	consumed      map[string]bool
}

type fakeUserRepository struct {
	repositories.UserRepository
	users map[string]*models.User
}

func (f *fakeUserRepository) FindById(id string) (*models.User, error) {
	if user, ok := f.users[id]; ok {
		return user, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (f *fakeRefreshTokenRepository) Create(token *models.RefreshToken) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, token)
	return nil
}

func (f *fakeRefreshTokenRepository) FindByHash(hash string) (*models.RefreshToken, error) {
	if token, ok := f.byHash[hash]; ok {
		return token, nil
	}
	return nil, repositories.ErrRefreshTokenInvalid
}

func (f *fakeRefreshTokenRepository) Consume(hash string, now time.Time) (*models.RefreshToken, error) {
	token, ok := f.byHash[hash]
	if !ok {
		return nil, repositories.ErrRefreshTokenInvalid
	}
	if f.consumed[hash] {
		return token, repositories.ErrRefreshTokenReused
	}
	f.consumed[hash] = true
	return token, nil
}

func (f *fakeRefreshTokenRepository) RevokeFamily(familyId primitive.ObjectID, now time.Time) error {
	f.revokedFamily = familyId
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

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	family := primitive.NewObjectID()
	repo := &fakeRefreshTokenRepository{byHash: map[string]*models.RefreshToken{
		utils.HashRefreshToken("known-token"): {FamilyId: family},
	}}
	router := gin.New()
	router.POST("/auth/logout", (&AuthController{refreshTokenRepository: repo}).Logout)

	cases := []struct {
		name        string
		body        string
		want        int
		wantRevoked bool
	}{
		{"known token revokes its family", `{"refresh_token":"known-token"}`, http.StatusOK, true},
		{"unknown token still logs out", `{"refresh_token":"unknown-token"}`, http.StatusOK, false},
		{"missing token is rejected", `{}`, http.StatusBadRequest, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo.revokedFamily = primitive.NilObjectID
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(tc.body)))

			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
			if revoked := repo.revokedFamily == family; revoked != tc.wantRevoked {
				t.Errorf("family revoked = %v, want %v", revoked, tc.wantRevoked)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	useTestSigningKey(t)

	family := primitive.NewObjectID()
	verified := &models.User{Id: primitive.NewObjectID(), Role: []string{"user"}, IsVerified: true}
	unverified := &models.User{Id: primitive.NewObjectID(), Role: []string{"user"}}
	demoted := &models.User{Id: primitive.NewObjectID(), Role: []string{"admin"}, IsVerified: true}

	tokens := map[string]*models.User{
		"valid-token":      verified,
		"reused-token":     verified,
		"unverified-token": unverified,
		"demoted-token":    demoted,
		"deleted-token":    {Id: primitive.NewObjectID()},
	}

	cases := []struct {
		name        string
		body        string
		usedAgo     time.Duration
		want        int
		wantRevoked bool
	}{
		{"valid token rotates in the same family", `{"refresh_token":"valid-token"}`, 0, http.StatusOK, false},
		{"reused token revokes the family", `{"refresh_token":"reused-token"}`, time.Minute, http.StatusUnauthorized, true},
		{"token used by another tab just now conflicts", `{"refresh_token":"reused-token"}`, 2 * time.Second, http.StatusConflict, false},
		{"unknown token is rejected", `{"refresh_token":"unknown-token"}`, 0, http.StatusUnauthorized, false},
		{"unverified user is rejected", `{"refresh_token":"unverified-token"}`, 0, http.StatusUnauthorized, false},
		{"user without the role is rejected", `{"refresh_token":"demoted-token"}`, 0, http.StatusUnauthorized, false},
		{"deleted user is rejected", `{"refresh_token":"deleted-token"}`, 0, http.StatusUnauthorized, false},
		{"missing token is rejected", `{}`, 0, http.StatusBadRequest, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRefreshTokenRepository{byHash: map[string]*models.RefreshToken{}, consumed: map[string]bool{}}
			users := &fakeUserRepository{users: map[string]*models.User{
				verified.Id.Hex():   verified,
				unverified.Id.Hex(): unverified,
				demoted.Id.Hex():    demoted,
			}}
			for raw, user := range tokens {
				hash := utils.HashRefreshToken(raw)
				token := &models.RefreshToken{FamilyId: family, SubjectId: user.Id, Role: "user"}
				if tc.usedAgo > 0 {
					usedAt := time.Now().Add(-tc.usedAgo)
					token.RevokedAt = &usedAt
					repo.consumed[hash] = true
				}
				repo.byHash[hash] = token
			}

			router := gin.New()
			router.POST("/auth/refresh", (&AuthController{userRepository: users, refreshTokenRepository: repo}).Refresh)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(tc.body)))

			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
			if revoked := repo.revokedFamily == family; revoked != tc.wantRevoked {
				t.Errorf("family revoked = %v, want %v", revoked, tc.wantRevoked)
			}
			if tc.want != http.StatusOK {
				if len(repo.created) != 0 {
					t.Errorf("issued %d tokens on a rejected refresh", len(repo.created))
				}
				return
			}

			var body struct {
				Data models.SessionTokensResponse `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Data.AccessToken == "" || body.Data.RefreshToken == "" || body.Data.RefreshToken == "valid-token" {
				t.Errorf("unexpected tokens %+v", body.Data)
			}
			if len(repo.created) != 1 || repo.created[0].FamilyId != family || repo.created[0].SubjectId != verified.Id {
				t.Errorf("new token not stored in the same family: %+v", repo.created)
			}
		})
	}
}
