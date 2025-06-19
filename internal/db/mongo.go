package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoURI = "mongodb://localhost:27017" // URI สำหรับเชื่อมต่อ MongoDB
	DBName   = "authManagement"            // ชื่อฐานข้อมูลที่จะใช้
)

// ฟังก์ชัน InitMongo ใช้สำหรับเชื่อมต่อกับ MongoDB และส่งคืน client กับ collection ที่ต้องการ
func InitMongo() (*mongo.Client, *mongo.Collection, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // เพื่อให้ยกเลิก context เมื่อฟังก์ชันนี้ทำงานเสร็จ

	// สร้าง client เชื่อมต่อ MongoDB โดยใช้ URI ที่กำหนดไว้
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		return nil, nil, nil, err
	}

	// ตรวจสอบว่าการเชื่อมต่อยังใช้ได้โดยการ ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, nil, err
	}

	log.Println("Connected to MongoDB")

	// เลือกฐานข้อมูลตามชื่อ DBName
	db := client.Database(DBName)

	// เลือก collection
	usersCollection := db.Collection("users")
	blacklistCollection := db.Collection("blacklisted_tokens")

	// ส่งคืนค่าที่กำหนด
	return client, usersCollection, blacklistCollection, nil
}
