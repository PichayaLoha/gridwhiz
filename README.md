
# GridWhiz Candidate Assignment

โปรเจกต์นี้เป็น Authentication Microservice
## โครงสร้างโปรเจกต์
- `auth-microervice/` : สำหรับเก็บ protoc (Protocol Buffers compiler)
- `auth/` : จัดการ JWT token, สร้างและตรวจสอบ token
- `db/` : ตั้งค่าและเชื่อมต่อกับ MongoDB
- `model/` : สำหรับเก็บโครงสร้างข้อมูล
- `server/` : สำหรับเซ็ตอัพ gRPC server
- `service/` : บริการหลัก เช่น Register, Login, Logout, User CRUD
- `validation/` : สำหรับตรวจสอบข้อมูล
- `proto/` : สำหรับเก็บไฟล์ .proto สำหรับ gRPC service และ message definitions

## ฟังก์ชันหลัก
- `Register` : ลงทะเบียนผู้ใช้ใหม่ พร้อมตรวจสอบข้อมูล
- `Login` : เข้าสู่ระบบ ตรวจสอบผู้ใช้และรหัสผ่าน, สร้าง JWT token และเก็บใน Redis
- `Logout` : ออกจากระบบ บล็อก token ปัจจุบันและลบจาก Redis
- `ListUsers` : ดึงค่าข้อมูลผู้ใช้ การทำPagination และการกำหนดสิทธิ์การเข้าถึง
- `GetUserById` : ดึงค่าข้อมูลผู้ใช้ตามไอดี
- `UpdateUser` : อัปเดตข้อมูลผู้ใช้
- `DeleteUser` : ลบข้อมูลผู้ใช้ (soft delete)
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

## ข้อมูลเพิ่มเติม
- JWT token หมดอายุทุก 5 นาที
- ต้องใช้ Docker Desktop ในการรัน Redis
- Redis ใช้เก็บ active token และนับ login attempts สำหรับ rate limiting