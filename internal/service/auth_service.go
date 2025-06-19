package service

import (
	pb "auth-microservice/auth-microservice/proto"
	"context"
	"fmt"
	"log"
	"time"

	"auth-microservice/internal/auth"
	models "auth-microservice/internal/model"
	"auth-microservice/internal/validation"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {
	// ตรวจสอบความถูกต้องของอีเมล (เช่น ซ้ำกับผู้ใช้อื่นหรือไม่)
	if err := validation.ValidateEmail(in.GetEmail(), ctx, s.UserCollection); err != nil {
		return nil, err
	}

	// ตรวจสอบความถูกต้องของ username
	if err := validation.ValidateUsername(in.GetUsername(), ctx, s.UserCollection); err != nil {
		return nil, err
	}

	// ตรวจสอบความแข็งแรงของรหัสผ่านและเข้ารหัสด้วย bcrypt
	hashedPassword, err1 := validation.ValidatePassword(in.GetPassword())
	if err1 != nil {
		return nil, err1
	}

	// สร้าง user document เตรียมสำหรับ insert ลง MongoDB
	user := map[string]interface{}{
		"email":     in.GetEmail(),
		"username":  in.GetUsername(),
		"password":  string(hashedPassword),
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
		"deleted":   false,
		"deletedAt": nil,
		"role":      in.GetRole(),
	}

	// บันทึกลงใน collection ของผู้ใช้
	_, err := s.UserCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	// ส่ง response กลับไปยัง client
	return &pb.RegisterReply{
		Email:     in.GetEmail(),
		Username:  in.GetUsername(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	// สร้าง filter เพื่อค้นหาผู้ใช้จาก email
	filter := bson.M{"email": in.GetEmail()}

	var user map[string]interface{}

	// ค้นหา user จาก MongoDB
	err := s.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้ที่มีอีเมลนี้")
	}

	// ตรวจสอบว่าเกิน rate limit หรือไม่
	isLimited, err := s.isRateLimited(ctx, in.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "ไม่สามารถตรวจสอบ Rate Limit ได้")
	}
	if isLimited {
		return nil, status.Error(codes.ResourceExhausted, "คุณพยายามเข้าสู่ระบบบ่อยเกินไป กรุณารอ 1 นาที")
	}

	// ตรวจสอบรหัสผ่านว่าตรงกับที่เก็บไว้หรือไม่
	hashedPassword := user["password"].(string)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(in.GetPassword()))
	if err != nil {
		// ถ้ารหัสผ่านผิด ก็ยังคงเพิ่ม count ให้ rate limit
		s.isRateLimited(ctx, in.GetEmail()) // เพิ่มการนับ rate limit เมื่อใส่รหัสผิดด้วย
		return nil, status.Error(codes.Unauthenticated, "รหัสผ่านไม่ถูกต้อง")
	}

	// ตรวจสอบและทำให้โทเค็นเก่าใช้งานไม่ได้
	userEmail := in.GetEmail()
	redisKey := fmt.Sprintf("active_token:%s", userEmail)

	oldToken, err := s.Redis.Get(ctx, redisKey).Result()
	if err == nil && oldToken != "" {
		// ถ้ามี token เก่าอยู่, ให้เพิ่มเข้า Blacklist
		blacklistedToken := models.BlacklistedToken{
			Token:     oldToken,
			ExpiresAt: time.Now(), // หมดอายุทันที
		}
		_, err := s.BlacklistCollection.InsertOne(ctx, blacklistedToken)
		if err != nil {
			return nil, status.Error(codes.Internal, "ไม่สามารถเพิ่ม token เข้า blacklisted ได้")
		}
	}

	// สร้าง JWT Token
	token, err := auth.GenerateJWT(in.GetEmail(), user["role"].(string))
	if err != nil {
		return nil, status.Error(codes.Internal, "เจอข้อผิดพลาดในการสร้างโทเค็น")
	}

	// บันทึก token ใหม่ลง Redis พร้อมระบุเวลา expiration
	jwtExpiration := time.Hour * 24
	err = s.Redis.Set(ctx, redisKey, token, jwtExpiration).Err()
	if err != nil {
		log.Printf("Could not set active token in Redis for user %s: %v", userEmail, err)
	}

	// ส่งข้อมูลกลับไปยัง client
	return &pb.LoginReply{
		Email:    user["email"].(string),
		Username: user["username"].(string),
		Token:    token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutReply, error) {
	//ดึง token มาใช้ในการ logout
	tokenStr := in.GetToken()

	if tokenStr == "" {
		return nil, status.Error(codes.InvalidArgument, "ต้องระบุ token")
	}

	// ตรวจสอบว่า token ถูก blacklist แล้วหรือยัง
	isBlacklisted, err := s.IsTokenBlacklisted(ctx, tokenStr)
	if err != nil {
		return nil, status.Error(codes.Internal, "เจอข้อผิดพลาดในการตรวจสอบโทเค็น")
	}
	if isBlacklisted {
		return nil, status.Error(codes.FailedPrecondition, "โทเค็นนี้ถูกบล็อกแล้ว")
	}

	// ตรวจสอบ JWT token ว่าถูกต้อง
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("วิธีการเซ็นชื่อโทเค็นไม่ถูกต้อง:: %v", token.Header["alg"])
		}
		return auth.JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "โทเค็นไม่ถูกต้องหรือหมดอายุ")
	}

	// ดึง email ของผู้ใช้
	var userEmail string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims["email"].(string); ok {
			userEmail = email
		}
	}

	if userEmail == "" {
		return nil, status.Error(codes.InvalidArgument, "ไม่สามารถระบุผู้ใช้จากโทเค็นได้")
	}

	// หาวันหมดอายุของ token เพื่อใส่ลง blacklist
	exp, err := auth.GetTokenExpiration(tokenStr)
	if err != nil {
		exp = time.Now().Add(24 * time.Hour) // กำหนด default expiration
	}

	// เพิ่ม token เข้า blacklist
	blacklistedToken := models.BlacklistedToken{
		Token:     tokenStr,
		ExpiresAt: exp,
	}

	// ลบ token ออกจาก Redis
	_, err = s.BlacklistCollection.InsertOne(ctx, blacklistedToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "ไม่สามารถบล็อกโทเค็นได้")
	}

	// ลบ Active Token ออกจาก Redis
	redisKey := fmt.Sprintf("active_token:%s", userEmail)
	if err := s.Redis.Del(ctx, redisKey).Err(); err != nil {
		log.Printf("Could not delete active token from Redis for user %s: %v", userEmail, err)
	}

	// ส่งข้อความว่า logout สำเร็จ
	return &pb.LogoutReply{
		Message: "ออกจากระบบสำเร็จ",
	}, nil
}
