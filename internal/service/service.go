package service

import (
	pb "auth-microservice/auth-microservice/proto"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// ฝัง default implementation เข้าไปใน struct ของเรา
type AuthService struct {
	UserCollection                    *mongo.Collection // MongoDB collection สำหรับเก็บข้อมูลผู้ใช้
	BlacklistCollection               *mongo.Collection // MongoDB collection สำหรับเก็บ token ที่ถูก blacklist
	Redis                             *redis.Client     // Redis client สำหรับใช้เก็บข้อมูลชั่วคราว เช่น rate limit และ token
	pb.UnimplementedAuthServiceServer                   // ฝัง default implementation ของ AuthService (จาก gRPC proto)
}

// สร้างอินสแตนซ์ของ AuthService พร้อมกำหนด collection และ redis client
func NewAuthService(userCol *mongo.Collection, blacklistCol *mongo.Collection, rdb *redis.Client) *AuthService { //dependecy injection
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

// สร้างอินสแตนซ์ของ UserService
func NewUserService(col *mongo.Collection) *UserService {
	return &UserService{UserCollection: col}
}
