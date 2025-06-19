package models

import "time"

type BlacklistedToken struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expires_at"`
}
