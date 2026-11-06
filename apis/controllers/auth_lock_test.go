package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils"
)

const testMaxFailures = 5

type fakeAuthAttemptRepository struct {
	repositories.AuthAttemptRepository
	failures map[string]int
}

func newFakeAuthAttempts() *fakeAuthAttemptRepository {
	return &fakeAuthAttemptRepository{failures: map[string]int{}}
}

func (f *fakeAuthAttemptRepository) Wait(key string, now time.Time) (time.Duration, error) {
	if f.failures[key] >= testMaxFailures {
		return 15 * time.Minute, nil
	}
	return 0, nil
}

func (f *fakeAuthAttemptRepository) RecordFailure(key string, now time.Time) error {
	f.failures[key]++
	return nil
}

func (f *fakeAuthAttemptRepository) Reset(key string) error {
	delete(f.failures, key)
	return nil
}

type fakeLoginUserRepository struct {
	repositories.UserRepository
	user *models.User
}

func (f *fakeLoginUserRepository) FindByEmail(email string) (*models.User, error) {
	if f.user != nil && f.user.Email == email {
		return f.user, nil
	}
	return nil, mongo.ErrNoDocuments
}

type fakeLoginAdminRepository struct {
	repositories.AdminRepository
	admin *models.Admin
}

func (f *fakeLoginAdminRepository) FindByEmail(email string) (*models.Admin, error) {
	if f.admin != nil && f.admin.Email == email {
		return f.admin, nil
	}
	return nil, mongo.ErrNoDocuments
}

func postLogin(handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	router := gin.New()
	router.POST("/login", handler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestAuthRetryAfterSeconds(t *testing.T) {
	cases := []struct {
		wait time.Duration
		want int
	}{
		{time.Second, 1},
		{1500 * time.Millisecond, 2},
		{15 * time.Minute, 900},
	}
	for _, tc := range cases {
		if got := authRetryAfterSeconds(tc.wait); got != tc.want {
			t.Errorf("authRetryAfterSeconds(%v) = %d, want %d", tc.wait, got, tc.want)
		}
	}
}

func TestSignInUserLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	useTestSigningKey(t)

	const password = "Correct-password-1"
	hashed, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	newController := func(attempts *fakeAuthAttemptRepository) *AuthController {
		return &AuthController{
			userRepository: &fakeLoginUserRepository{user: &models.User{
				Id:         primitive.NewObjectID(),
				Email:      "a@b.com",
				Password:   hashed,
				Role:       []string{"user"},
				IsVerified: true,
			}},
			refreshTokenRepository: &fakeRefreshTokenRepository{},
			authAttemptRepository:  attempts,
		}
	}
	login := func(ac *AuthController, email, pass string) *httptest.ResponseRecorder {
		body := `{"email":"` + email + `","password":"` + pass + `","role":"user"}`
		return postLogin(ac.SignInUser, body)
	}

	t.Run("locks after repeated wrong passwords even with the right one", func(t *testing.T) {
		ac := newController(newFakeAuthAttempts())
		for i := 0; i < testMaxFailures; i++ {
			if w := login(ac, "a@b.com", "wrong"); w.Code != http.StatusBadRequest {
				t.Fatalf("attempt %d status = %d, want 400", i+1, w.Code)
			}
		}

		w := login(ac, "a@b.com", password)
		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want 429", w.Code)
		}
		var got struct {
			Data models.RetryAfter `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Data.RetryAfterSeconds != 900 {
			t.Errorf("retry_after_seconds = %d, want 900", got.Data.RetryAfterSeconds)
		}
	})

	t.Run("an unknown email locks the same way", func(t *testing.T) {
		ac := newController(newFakeAuthAttempts())
		for i := 0; i < testMaxFailures; i++ {
			login(ac, "nobody@b.com", "wrong")
		}
		if w := login(ac, "nobody@b.com", "wrong"); w.Code != http.StatusTooManyRequests {
			t.Errorf("status = %d, want 429", w.Code)
		}
	})

	t.Run("a correct password clears the count", func(t *testing.T) {
		attempts := newFakeAuthAttempts()
		ac := newController(attempts)
		login(ac, "a@b.com", "wrong")
		if w := login(ac, "a@b.com", password); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if got := attempts.failures["login:a@b.com"]; got != 0 {
			t.Errorf("failures = %d, want 0", got)
		}
	})
}

func TestSignInAdminLock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hashed, err := utils.HashPassword("Correct-password-1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	attempts := newFakeAuthAttempts()
	ac := &AuthController{
		adminRepository:       &fakeLoginAdminRepository{admin: &models.Admin{Email: "admin@b.com", Password: hashed}},
		authAttemptRepository: attempts,
	}

	for i := 0; i < testMaxFailures; i++ {
		postLogin(ac.SignInAdmin, `{"email":"admin@b.com","password":"wrong"}`)
	}
	if got := attempts.failures["login-admin:admin@b.com"]; got != testMaxFailures {
		t.Errorf("failures = %d, want %d", got, testMaxFailures)
	}
	w := postLogin(ac.SignInAdmin, `{"email":"admin@b.com","password":"Correct-password-1"}`)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}

func TestChangePasswordLock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hashed, err := utils.HashPassword("Current-password-1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := primitive.NewObjectID()
	attempts := newFakeAuthAttempts()
	ac := &AuthController{
		userRepository: &fakeUserRepository{users: map[string]*models.User{
			userID.Hex(): {Id: userID, Password: hashed},
		}},
		authAttemptRepository: attempts,
	}
	change := func(current string) *httptest.ResponseRecorder {
		router := gin.New()
		router.PATCH("/changepassword/:id", ac.ChangePassword)
		w := httptest.NewRecorder()
		body := `{"current_password":"` + current + `","new_password":"New-password-1"}`
		req := httptest.NewRequest(http.MethodPatch, "/changepassword/"+userID.Hex(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w
	}

	for i := 0; i < testMaxFailures; i++ {
		if w := change("wrong"); w.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d status = %d, want 400", i+1, w.Code)
		}
	}
	if got := attempts.failures["change-password:"+userID.Hex()]; got != testMaxFailures {
		t.Errorf("failures = %d, want %d", got, testMaxFailures)
	}
	if w := change("Current-password-1"); w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}
