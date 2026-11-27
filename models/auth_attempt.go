package models

import "time"

type AuthAttempt struct {
	Key         string    `bson:"key"`
	Attempts    int       `bson:"attempts"`
	LockedUntil time.Time `bson:"locked_until,omitempty"`
	ExpiresAt   time.Time `bson:"expires_at"`
}
