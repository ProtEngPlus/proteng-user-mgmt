package controllers

import (
	"testing"
	"time"
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
