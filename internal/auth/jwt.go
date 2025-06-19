package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// secret key สำหรับใช้เซ็นและตรวจสอบ JWT token
var JwtSecret = []byte("secret-key") // ใช้ตัวนี้ที่เดียวพอ

// สร้าง JWT token
func GenerateJWT(email string, role string) (string, error) {

	// สร้าง claims สำหรับใส่ข้อมูลใน token
	claims := jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Minute * 5).Unix(),
	}

	//// สร้าง token ใหม่โดยใช้ HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// ตรวจสอบและดึงเวลาหมดอายุของ token
func GetTokenExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return time.Time{}, err
	}

	// ถ้าแปลง claims ได้สำเร็จ ให้ดึงข้อมูล exp ออกมา
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0), nil
		}
	}

	return time.Time{}, errors.New("ไม่พบเวลาหมดอายุในโทเค็น")
}

// แปลง token string เป็น claims map[string]interface{} เพื่อดึงข้อมูลใน token
func ParseToken(tokenStr string) (map[string]interface{}, error) {
	// แปลง token string และตรวจสอบความถูกต้อง
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	// แปลง claims เป็น map และคืนค่า
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return nil, errors.New("ไม่สามารถแปลง claims เป็น map ได้")
}
