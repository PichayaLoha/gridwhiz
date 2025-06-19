package service

import (
	"context"
	"fmt"
	"time"
)

func (s *AuthService) isRateLimited(ctx context.Context, email string) (bool, error) {
	// สร้าง key สำหรับ Redis โดยอิงจาก email ผู้ใช้
	key := fmt.Sprintf("login_attempt:%s", email)

	// เพิ่มจำนวนการพยายาม login ของ email นี้ใน Redis ทีละ 1
	attempts, err := s.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	//เริ่มจับเวลานับ 1 นาที ตั้งแต่พยายาม login ครั้งแรก
	if attempts == 1 {
		// กำหนดให้ key นี้หมดอายุใน 1 นาที (rate limit window 1 นาที)
		s.Redis.Expire(ctx, key, time.Minute)
	}

	// ถ้าผู้ใช้พยายาม login เกิน 5 ครั้งใน 1 นาที ให้ถือว่าถูกจำกัด
	return attempts > 5, nil
}
