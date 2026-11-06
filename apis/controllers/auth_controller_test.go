package controllers

import (
	"testing"
	"time"

	"github.com/protengplus/proteng-user-mgmt/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func TestVerificationResendWait(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	// expireIfSent returns the token expiry stored when a token was sent `ago` before now.
	expireIfSent := func(ago time.Duration) time.Time {
		return now.Add(-ago).Add(verificationTokenDaysTTL * 24 * time.Hour)
	}

	cases := []struct {
		name   string
		expire time.Time
		want   time.Duration
	}{
		{"sent just now", expireIfSent(0), 60 * time.Second},
		{"sent 45s ago", expireIfSent(45 * time.Second), 15 * time.Second},
		{"sent exactly 60s ago", expireIfSent(60 * time.Second), 0},
		{"sent an hour ago", expireIfSent(time.Hour), 0},
		{"no token stored", time.Time{}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := verificationResendWait(tc.expire, now); got != tc.want {
				t.Errorf("verificationResendWait() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPasswordResetResendWait(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	expireIfSent := func(ago time.Duration) time.Time {
		return now.Add(-ago).Add(passwordResetTokenMinutesTTL * time.Minute)
	}

	cases := []struct {
		name   string
		expire time.Time
		want   time.Duration
	}{
		{"sent just now", expireIfSent(0), 60 * time.Second},
		{"sent 45s ago", expireIfSent(45 * time.Second), 15 * time.Second},
		{"sent exactly 60s ago", expireIfSent(60 * time.Second), 0},
		{"sent an hour ago", expireIfSent(time.Hour), 0},
		{"no token stored", time.Time{}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := passwordResetResendWait(tc.expire, now); got != tc.want {
				t.Errorf("passwordResetResendWait() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPasswordResetTokenQuery(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	query := passwordResetTokenQuery("raw-token", now)

	if len(query) != 2 {
		t.Fatalf("query has %d conditions, want 2", len(query))
	}
	if query[0].Key != "passwordResetToken" || query[0].Value != utils.HashResetToken("raw-token") {
		t.Errorf("token condition = %v, want the hashed token", query[0])
	}
	if query[0].Value == "raw-token" {
		t.Error("query must not use the raw token")
	}
	expire, ok := query[1].Value.(bson.D)
	if query[1].Key != "passwordResetTokenExpire" || !ok || len(expire) != 1 {
		t.Fatalf("expiry condition = %v", query[1])
	}
	if expire[0].Key != "$gt" || expire[0].Value != now {
		t.Errorf("expiry condition = %v, want $gt now", expire[0])
	}
}
