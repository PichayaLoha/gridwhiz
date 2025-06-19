package validation

import (
	"context"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validate Email
func ValidateEmail(email string, ctx context.Context, userCollection *mongo.Collection) error {
	filterEmail := bson.M{"email": email}
	count, err := userCollection.CountDocuments(ctx, filterEmail)
	re := regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if err != nil {
		return err
	}

	if count > 0 {
		return status.Error(codes.AlreadyExists, "อีเมลถูกใช้งานแล้ว")
	}

	if strings.TrimSpace(email) == "" {
		return status.Error(codes.InvalidArgument, "อีเมลไม่สามารถเว้นว่างได้")
	}

	if len(email) > 254 {
		return status.Error(codes.InvalidArgument, "อีเมลต้องมีความยาวไม่เกิน 254 ตัวอักษร")
	}

	if !re.MatchString(email) {
		return status.Error(codes.InvalidArgument, "รูปแบบอีเมลไม่ถูกต้อง")
	}

	return nil
}

// Validate Password
func ValidatePassword(password string) (string, error) {
	if len(password) < 6 {
		return "", status.Error(codes.InvalidArgument, "พาสเวิร์ดต้องมีความยาวอย่างน้อย 6 ตัวอักษร")
	}

	hasNumber := regexp.MustCompile("[0-9]").MatchString(password)
	hasUpper := regexp.MustCompile("[A-Z]").MatchString(password)
	hasLower := regexp.MustCompile("[a-z]").MatchString(password)

	if !hasNumber || !hasUpper || !hasLower {
		return "", status.Error(codes.InvalidArgument, "พาสเวิร์ดต้องมีตัวเลข ตัวพิมพ์ใหญ่ และตัวพิมพ์เล็กอย่างน้อย 1 ตัว")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// Validate Username
func ValidateUsername(username string, ctx context.Context, userCollection *mongo.Collection) error {
	filterUsername := bson.M{"username": username}
	count, err := userCollection.CountDocuments(ctx, filterUsername)
	if err != nil {
		return err
	}

	if count > 0 {
		return status.Error(codes.AlreadyExists, "ชื่อผู้ใช้ถูกใช้งานแล้ว")
	}

	if len(username) < 3 || len(username) > 20 {
		return status.Error(codes.InvalidArgument, "ชื่อผู้ใช้ต้องมีความยาวระหว่าง 3 ถึง 20 ตัวอักษร")
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !re.MatchString(username) {
		return status.Error(codes.InvalidArgument, "ห้ามใช้ตัวอักษรพิเศษในชื่อผู้ใช้")
	}

	return nil
}
