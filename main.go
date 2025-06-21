package main

import (
	"log"

	"auth-microservice/internal/server"
)

func main() {
	// =================เริ่มการทำงาน=================
	log.Println("Starting Auth Microservice gRPC server...")
	if err := server.RunGRPCServer(); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

// func main() {
// 	log.Println("Starting Auth Microservice gRPC server...")

// 	authService, err := server.RunGRPCServer()
// 	if err != nil {
// 		log.Fatalf("Failed to start gRPC server: %v", err)
// 	}

// 	// รอ server ตั้งต้นเสร็จ
// 	time.Sleep(2 * time.Second)

// 	// เรียก Bulk Register 100,000 users
// 	go func() {
// 		log.Println("Starting bulk registration...")
// 		err := authService.RegisterBulkUsers(context.Background(), 100000)
// 		if err != nil {
// 			log.Printf("Bulk registration failed: %v", err)
// 		} else {
// 			log.Println("Bulk registration completed successfully.")
// 		}
// 	}()

// 	// กัน main ออกจากโปรแกรม
// 	select {}
// }
