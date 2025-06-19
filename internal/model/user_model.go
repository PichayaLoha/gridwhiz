package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var User struct {
	ID        primitive.ObjectID `bson:"_id"`
	Email     string             `bson:"email"`
	Username  string             `bson:"username"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}
