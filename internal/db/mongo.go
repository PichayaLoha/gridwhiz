package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoURI = "mongodb://localhost:27017"
	DBName   = "authManagement"
)

func InitMongo() (*mongo.Client, *mongo.Collection, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		return nil, nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, nil, err
	}

	log.Println("Connected to MongoDB")

	db := client.Database(DBName)
	usersCollection := db.Collection("users")
	blacklistCollection := db.Collection("blacklisted_tokens")

	return client, usersCollection, blacklistCollection, nil
}
