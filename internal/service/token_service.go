package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	filter := bson.M{"token": token}
	count, err := s.BlacklistCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
