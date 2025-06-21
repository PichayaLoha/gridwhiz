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

const grpcPort = ":50051" // พอร์ตที่ gRPC server จะรับการเชื่อมต่อ

func RunGRPCServer() error {
	//  ===== เชื่อมต่อ MongoDB  =====
	client, userCollection, blacklistCollection, err := db.InitMongo()
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background()) // ปิดการเชื่อมต่อเมื่อ server หยุดทำงาน

	// ===== เชื่อมต่อกับ Redis =====
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// ===== กำหนดพอร์ต gRPC listener  =====
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	// ===== สร้าง gRPC Server =====
	grpcServer := grpc.NewServer()

	//===== สร้าง service instances และ inject dependencies =====
	authService := service.NewAuthService(userCollection, blacklistCollection, rdb)
	userService := service.NewUserService(userCollection)

	// ===== Register gRPC service =====
	pb.RegisterAuthServiceServer(grpcServer, authService)
	pb.RegisterUserServiceServer(grpcServer, userService)
	log.Printf("gRPC server listening on %s", grpcPort)

	// เริ่มรัน gRPC
	return grpcServer.Serve(lis)
}

// แก้จาก `error` เป็น `(*service.AuthService, error)`
// func RunGRPCServer() (*service.AuthService, error) {
// 	// ===== เชื่อมต่อ MongoDB =====
// 	_, userCollection, blacklistCollection, err := db.InitMongo()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// ไม่ defer Disconnect ที่นี่ เพราะเราจะ return service ไปใช้งานต่อ

// 	// ===== เชื่อมต่อกับ Redis =====
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "",
// 		DB:       0,
// 	})

// 	// ===== สร้าง service instances =====
// 	authService := service.NewAuthService(userCollection, blacklistCollection, rdb)
// 	userService := service.NewUserService(userCollection)

// 	// ===== Register gRPC =====
// 	lis, err := net.Listen("tcp", grpcPort)
// 	if err != nil {
// 		return nil, err
// 	}

// 	grpcServer := grpc.NewServer()
// 	pb.RegisterAuthServiceServer(grpcServer, authService)
// 	pb.RegisterUserServiceServer(grpcServer, userService)

// 	log.Printf("gRPC server listening on %s", grpcPort)

// 	// รัน server ใน goroutine เพื่อไม่บล็อก
// 	go func() {
// 		if err := grpcServer.Serve(lis); err != nil {
// 			log.Fatalf("gRPC server failed: %v", err)
// 		}
// 	}()

// 	// return authService เพื่อให้ main.go เรียกใช้ได้
// 	return authService, nil
// }
