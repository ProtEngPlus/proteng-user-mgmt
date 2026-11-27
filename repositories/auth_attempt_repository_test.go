package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
)

func TestAuthLockRemaining(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		lockedUntil time.Time
		want        time.Duration
	}{
		{"never locked", time.Time{}, 0},
		{"lock already passed", now.Add(-time.Second), 0},
		{"lock ends right now", now, 0},
		{"still locked", now.Add(10 * time.Minute), 10 * time.Minute},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := authLockRemaining(tc.lockedUntil, now); got != tc.want {
				t.Errorf("authLockRemaining() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestShouldLockAuth(t *testing.T) {
	cases := []struct {
		attempts int
		want     bool
	}{
		{1, false},
		{maxAuthFailures - 1, false},
		{maxAuthFailures, true},
		{maxAuthFailures + 1, true},
	}

	for _, tc := range cases {
		if got := shouldLockAuth(tc.attempts); got != tc.want {
			t.Errorf("shouldLockAuth(%d) = %v, want %v", tc.attempts, got, tc.want)
		}
	}
}

func newTestAuthAttemptRepository(t *testing.T) AuthAttemptRepository {
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
		database.GetCollection("auth_attempts").Drop(context.Background())
	})

	repo := NewAuthAttemptRepository()
	if err := repo.EnsureIndexes(); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	return repo
}

func TestAuthAttemptRepository(t *testing.T) {
	repo := newTestAuthAttemptRepository(t)
	now := time.Now()

	fail := func(t *testing.T, key string, times int) {
		t.Helper()
		for i := 0; i < times; i++ {
			if err := repo.RecordFailure(key, now); err != nil {
				t.Fatalf("record failure: %v", err)
			}
		}
	}
	wait := func(t *testing.T, key string, at time.Time) time.Duration {
		t.Helper()
		got, err := repo.Wait(key, at)
		if err != nil {
			t.Fatalf("wait: %v", err)
		}
		return got
	}

	t.Run("not locked before the limit", func(t *testing.T) {
		fail(t, "below-limit", maxAuthFailures-1)
		if got := wait(t, "below-limit", now); got != 0 {
			t.Errorf("wait = %v, want 0", got)
		}
	})

	t.Run("locked at the limit", func(t *testing.T) {
		fail(t, "at-limit", maxAuthFailures)
		if got := wait(t, "at-limit", now); got <= 0 || got > authLockDuration {
			t.Errorf("wait = %v, want up to %v", got, authLockDuration)
		}
	})

	t.Run("count starts over after the lock ends", func(t *testing.T) {
		fail(t, "lock-ended", maxAuthFailures)
		if got := wait(t, "lock-ended", now.Add(authLockDuration+time.Second)); got != 0 {
			t.Fatalf("wait after lock = %v, want 0", got)
		}
		fail(t, "lock-ended", 1)
		if got := wait(t, "lock-ended", now); got != 0 {
			t.Errorf("wait after one new failure = %v, want 0", got)
		}
	})

	t.Run("reset clears the count", func(t *testing.T) {
		fail(t, "reset", maxAuthFailures-1)
		if err := repo.Reset("reset"); err != nil {
			t.Fatalf("reset: %v", err)
		}
		fail(t, "reset", 1)
		if got := wait(t, "reset", now); got != 0 {
			t.Errorf("wait = %v, want 0", got)
		}
	})
}
