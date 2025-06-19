package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	// สร้าง filter สำหรับค้นหา token ที่ถูก block
	filter := bson.M{"token": token}
	// ใช้ MongoDB นับจำนวน token ที่ตรงกับ filter
	// ถ้ามีแสดงว่า token นี้ถูกเพิ่มใน blacklist แล้ว
	count, err := s.BlacklistCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
