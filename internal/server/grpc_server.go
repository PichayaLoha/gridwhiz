package server

import (
	"context"
	"log"
	"net"

	"auth-microservice/internal/db"
	"auth-microservice/internal/service"

	pb "auth-microservice/auth-microservice/proto"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

const grpcPort = ":50051"

func RunGRPCServer() error {
	// เชื่อมต่อ MongoDB
	client, userCollection, blacklistCollection, err := db.InitMongo()
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// เชื่อมต่อ Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // หรือใช้ ENV config
		Password: "",
		DB:       0,
	})
	// เตรียม gRPC listener
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	authService := service.NewAuthService(userCollection, blacklistCollection, rdb)
	userService := service.NewUserService(userCollection)

	// Register gRPC service
	pb.RegisterAuthServiceServer(grpcServer, authService)
	pb.RegisterUserServiceServer(grpcServer, userService)
	log.Printf("gRPC server listening on %s", grpcPort)

	// เริ่มรัน gRPC
	return grpcServer.Serve(lis)
}
