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
