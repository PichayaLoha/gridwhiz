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
	if err := validation.ValidateEmail(in.GetEmail(), ctx, s.UserCollection); err != nil {
		return nil, err
	}

	if err := validation.ValidateUsername(in.GetUsername(), ctx, s.UserCollection); err != nil {
		return nil, err
	}

	hashedPassword, err1 := validation.ValidatePassword(in.GetPassword())
	if err1 != nil {
		return nil, err1
	}

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

	_, err := s.UserCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterReply{
		Email:     in.GetEmail(),
		Username:  in.GetUsername(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	filter := bson.M{"email": in.GetEmail()}

	var user map[string]interface{}

	err := s.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้ที่มีอีเมลนี้")
	}

	isLimited, err := s.isRateLimited(ctx, in.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "ไม่สามารถตรวจสอบ Rate Limit ได้")
	}
	if isLimited {
		return nil, status.Error(codes.ResourceExhausted, "คุณพยายามเข้าสู่ระบบบ่อยเกินไป กรุณารอ 1 นาที")
	}

	hashedPassword := user["password"].(string)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(in.GetPassword()))
	if err != nil {
		s.isRateLimited(ctx, in.GetEmail()) // เพิ่มการนับ rate limit เมื่อใส่รหัสผิดด้วย
		return nil, status.Error(codes.Unauthenticated, "รหัสผ่านไม่ถูกต้อง")
	}

	// ตรวจสอบและทำให้โทเค็นเก่าใช้งานไม่ได้
	userEmail := in.GetEmail()
	redisKey := fmt.Sprintf("active_token:%s", userEmail)

	oldToken, err := s.Redis.Get(ctx, redisKey).Result()
	if err == nil && oldToken != "" {
		// ถ้ามี token เก่าอยู่, ให้เพิ่มเข้า Blacklist
		// log.Printf("ผู้ใช้ %s ได้เข้าสู่ระบบอยู่แล้ว กำลังยกเลิกโทเค็นเดิม", userEmail)
		blacklistedToken := models.BlacklistedToken{
			Token:     oldToken,
			ExpiresAt: time.Now(), // หมดอายุทันที
		}
		_, err := s.BlacklistCollection.InsertOne(ctx, blacklistedToken)
		if err != nil {

			// log.Printf("Could not blacklist old token for user %s: %v", userEmail, err)
		}
	}

	// สร้าง JWT Token
	token, err := auth.GenerateJWT(in.GetEmail(), user["role"].(string))
	if err != nil {
		return nil, status.Error(codes.Internal, "เจอข้อผิดพลาดในการสร้างโทเค็น")
	}

	jwtExpiration := time.Hour * 24
	err = s.Redis.Set(ctx, redisKey, token, jwtExpiration).Err()
	if err != nil {
		log.Printf("Could not set active token in Redis for user %s: %v", userEmail, err)
	}

	fmt.Println("Generated token:", token)
	return &pb.LoginReply{
		Email:    user["email"].(string),
		Username: user["username"].(string),
		Token:    token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutReply, error) {
	log.Printf("Received logout request: %+v\n", in)
	tokenStr := in.GetToken()
	fmt.Println("Received token:", tokenStr)

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

	var userEmail string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims["email"].(string); ok {
			userEmail = email
		}
	}

	if userEmail == "" {
		return nil, status.Error(codes.InvalidArgument, "ไม่สามารถระบุผู้ใช้จากโทเค็นได้")
	}

	// ดึงเวลาหมดอายุของ token
	exp, err := auth.GetTokenExpiration(tokenStr)
	if err != nil {
		exp = time.Now().Add(24 * time.Hour) // กำหนด default expiration
	}

	blacklistedToken := models.BlacklistedToken{
		Token:     tokenStr,
		ExpiresAt: exp,
	}

	_, err = s.BlacklistCollection.InsertOne(ctx, blacklistedToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "ไม่สามารถบล็อกโทเค็นได้")
	}
	// ลบ Active Token ออกจาก Redis
	redisKey := fmt.Sprintf("active_token:%s", userEmail)
	if err := s.Redis.Del(ctx, redisKey).Err(); err != nil {
		log.Printf("Could not delete active token from Redis for user %s: %v", userEmail, err)
	}

	return &pb.LogoutReply{
		Message: "ออกจากระบบสำเร็จ",
	}, nil
}
