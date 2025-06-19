package service

import (
	"context"
	"fmt"
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
	"auth-microservice/internal/validation"
)

func (s *UserService) GetUserById(ctx context.Context, in *pb.UserIdRequest) (*pb.UserIdReply, error) {
	// แปลง string เป็น ObjectID
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ไอดีไม่ถูกต้อง")
	}

	filter := bson.M{
		"_id":     objID,
		"deleted": bson.M{"$ne": true}, // ดึงเฉพาะที่ยังไม่ถูกลบ != true
	}

	var user struct {
		ID        primitive.ObjectID `bson:"_id"`
		Email     string             `bson:"email"`
		Username  string             `bson:"username"`
		CreatedAt time.Time          `bson:"createdAt"`
		UpdatedAt time.Time          `bson:"updatedAt"`
	}

	err = s.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้")
	}

	return &pb.UserIdReply{
		Id:        user.ID.Hex(),
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	// แปลง id เป็น ObjectID
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ID ไม่ถูกต้อง")
	}
	if err := validation.ValidateEmail(in.GetEmail(), ctx, s.UserCollection); err != nil {
		return nil, err
	}
	if err := validation.ValidateUsername(in.GetUsername(), ctx, s.UserCollection); err != nil {
		return nil, err
	}
	update := bson.M{
		"$set": bson.M{
			"email":     in.GetEmail(),
			"username":  in.GetUsername(),
			"updatedAt": time.Now(),
		},
	}

	filter := bson.M{"_id": objID}

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
	objID, err := primitive.ObjectIDFromHex(in.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "ID ไม่ถูกต้อง")
	}

	filter := bson.M{"_id": objID, "deleted": bson.M{"$ne": true}} // ต้องไม่ถูกลบอยู่ก่อนแล้ว
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": time.Now(),
		},
	}

	result, err := s.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Error(codes.Internal, "เกิดข้อผิดพลาดในการลบผู้ใช้")
	}

	if result.MatchedCount == 0 {
		return nil, status.Error(codes.NotFound, "ไม่พบผู้ใช้ที่ต้องการลบหรือถูกลบไปแล้ว")
	}

	return &pb.DeleteUserReply{
		Message: "ลบข้อมูลผู้ใช้สำเร็จ (soft delete)",
	}, nil
}
func (s *UserService) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersReply, error) {
	// log.Println("ListUsers called with:", in)
	// log.Printf("ListUsers called: name=%s email=%s page=%d limit=%d", in.GetName(), in.GetEmail(), in.GetPage(), in.GetLimit())
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
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
	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// ตรวจสอบ role จาก claims map
	role, ok := claims["role"].(string)
	fmt.Println("User role from token:", role)
	if !ok || role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied: admin role required")
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

	// Pagination
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

	return &pb.ListUsersReply{
		Users: users,
		Total: int32(total),
	}, nil
}
