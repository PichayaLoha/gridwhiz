package models

import (
	"time"
)

type User struct {
	Email     string     `bson:"email"`
	Username  string     `bson:"username"`
	Password  string     `bson:"password"`
	CreatedAt time.Time  `bson:"createdAt"`
	UpdatedAt time.Time  `bson:"updatedAt"`
	Deleted   bool       `bson:"deleted"`
	DeletedAt *time.Time `bson:"deletedAt"`
	Role      string     `bson:"role"`
}

type BlacklistedToken struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expires_at"`
}
