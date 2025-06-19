package service

import (
	pb "auth-microservice/auth-microservice/proto"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	UserCollection      *mongo.Collection
	BlacklistCollection *mongo.Collection
	Redis               *redis.Client
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(userCol *mongo.Collection, blacklistCol *mongo.Collection, rdb *redis.Client) *AuthService {
	return &AuthService{
		UserCollection:      userCol,
		BlacklistCollection: blacklistCol,
		Redis:               rdb,
	}
}

type UserService struct {
	UserCollection *mongo.Collection
	pb.UnimplementedUserServiceServer
}

func NewUserService(col *mongo.Collection) *UserService {
	return &UserService{UserCollection: col}
}
