# GridWhiz Candidate Assignment
โปรเจกต์นี้เป็น Authentication Microservice
## การติดตั้งและรันโปรเจกต์
เปิดเทอร์มินัลในโฟลเดอร์โปรเจกต์ แล้วรันคำสั่ง:

ติดตั้ง gRPC framework สำหรับ Go
```
go get google.golang.org/grpc
go get google.golang.org/grpc/credentials/insecure
```
ติดตั้ง MongoDB driver สำหรับ Go
```
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
```
ติดตั้ง ฟังก์ชันเข้ารหัสรหัสผ่าน
```
go get golang.org/x/crypto/bcrypt
```
ติดตั้ง Redis
```
go get github.com/redis/go-redis/v9
```
ติดตั้ง JWT สำหรับ สร้างและตรวจสอบ JWT Token (JSON Web Token)
```
go get github.com/golang-jwt/jwt/v5
```
ติดตั้ง CLI ชื่อ ghz ที่ใช้ ทดสอบความเร็วและ performance ของ gRPC
```
go install github.com/bojand/ghz/cmd/ghz@latest
```

สั่งให้ Docker Compose สร้าง container และรันเบื้องหลัง(รัน Redis)
```
docker-compose up -d --build
```

สั่งรัน Go โปรแกรมหลัก
```
go run main.go
```
