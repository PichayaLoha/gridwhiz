package service

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "auth-microservice/auth-microservice/proto"
	"auth-microservice/internal/auth"
	models "auth-microservice/internal/model"
	"auth-microservice/internal/validation"
)

func (s *UserService) GetUserById(ctx context.Context, in *pb.UserIdRequest) (*pb.UserIdReply, error) {
	// แปลง string เป็น ObjectID
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ไอดีไม่ถูกต้อง")
	}

	// กำหนด filte
	filter := bson.M{
		"_id":     objID,
		"deleted": bson.M{"$ne": true}, // ดึงเฉพาะที่ยังไม่ถูกลบ != true
	}

	// ดึงข้อมูลผู้ใช้จาก MongoDB
	err = s.UserCollection.FindOne(ctx, filter).Decode(&models.User)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้")
	}

	// ส่งข้อมูลกลับในรูปแบบ protobuf
	return &pb.UserIdReply{
		Id:        models.User.ID.Hex(),
		Email:     models.User.Email,
		Username:  models.User.Username,
		CreatedAt: models.User.CreatedAt.Format(time.RFC3339),
		UpdatedAt: models.User.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	// แปลง id เป็น ObjectID
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ID ไม่ถูกต้อง")
	}

	// ตรวจสอบความถูกต้องของ username
	if err := validation.ValidateUsername(in.GetUsername(), ctx, s.UserCollection); err != nil {
		return nil, err
	}

	// กำหนดข้อมูลที่จะอัปเดต
	update := bson.M{
		"$set": bson.M{
			"username":  in.GetUsername(),
			"updatedAt": time.Now(),
		},
	}

	// อัปเดตข้อมูลใน MongoDB
	filter := bson.M{"_id": objID}

	// เช็คว่ามีผู้ใช้ตรงกับ id หรือไม่
	result, err := s.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Error(codes.Internal, "เกิดข้อผิดพลาดในการอัปเดตข้อมูล")
	}

	if result.MatchedCount == 0 {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้ที่ต้องการอัปเดต")
	}

	return &pb.UpdateUserReply{
		Message: "อัปเดตข้อมูลผู้ใช้สำเร็จ",
	}, nil
}
func (s *UserService) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	// แปลง id จาก string เป็น MongoDB ObjectID
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ID ไม่ถูกต้อง")
	}

	//filter
	filter := bson.M{"_id": objID, "deleted": bson.M{"$ne": true}} // ต้องไม่ถูกลบอยู่ก่อนแล้ว
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": time.Now(),
		},
	}

	// อัปเดตข้อมูล
	result, err := s.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Error(codes.Internal, "เกิดข้อผิดพลาดในการลบผู้ใช้")
	}

	// เช็คว่ามีผู้ใช้ตรงกับ id หรือไม่ หรือถูกลบไปแล้ว
	if result.MatchedCount == 0 {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้ที่ต้องการลบหรือถูกลบไปแล้ว")
	}

	return &pb.DeleteUserReply{
		Message: "ลบข้อมูลผู้ใช้สำเร็จ (soft delete)",
	}, nil
}
func (s *UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersReply, error) {
	// ดึง metadata จาก context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// ดึงค่า authorization header
	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not supplied")
	}

	// ดึง token จาก "Bearer <token>"
	tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if tokenStr == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization token is empty")
	}

	// เรียก ParseToken โดยส่ง token string
	// ตรวจสอบและแปลง token เป็น claims
	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// ตรวจสอบ role  == admin
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "สามารถดูได้เฉพาะ Admin เท่านั้น")
	}
	// สร้าง filter ค้นหา (ไม่รวมผู้ถูกลบ)
	filter := bson.M{"deleted": bson.M{"$ne": true}}

	// เงื่อนไขค้นหาด้วยชื่อ/อีเมล
	if in.GetName() != "" {
		filter["username"] = bson.M{"$regex": in.GetName(), "$options": "i"}
	}
	if in.GetEmail() != "" {
		filter["email"] = bson.M{"$regex": in.GetEmail(), "$options": "i"}
	}

	// กำหนดค่าการแบ่งหน้า
	page := in.GetPage()
	if page < 1 {
		page = 1
	}
	limit := in.GetLimit()
	if limit <= 0 {
		limit = 10
	}
	skip := (page - 1) * limit

	// ดึงข้อมูลจาก MongoDB
	cursor, err := s.UserCollection.Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(skip)))
	if err != nil {
		return nil, status.Error(codes.Internal, "เกิดข้อผิดพลาดในการค้นหา")
	}
	defer cursor.Close(ctx)

	// สร้างรายการผู้ใช้
	var users []*pb.UserItem
	for cursor.Next(ctx) {
		var u struct {
			ID        primitive.ObjectID `bson:"_id"`
			Email     string             `bson:"email"`
			Username  string             `bson:"username"`
			CreatedAt time.Time          `bson:"createdAt"`
		}
		if err := cursor.Decode(&u); err != nil {
			continue
		}
		users = append(users, &pb.UserItem{
			Id:        u.ID.Hex(),
			Email:     u.Email,
			Username:  u.Username,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		})
	}

	// นับจำนวนทั้งหมด
	total, err := s.UserCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, "ไม่สามารถนับผู้ใช้ทั้งหมดได้")
	}

	// ส่งข้อมูลกลับเป็น protobuf
	return &pb.ListUsersReply{
		Users: users,
		Total: int32(total),
	}, nil
}
