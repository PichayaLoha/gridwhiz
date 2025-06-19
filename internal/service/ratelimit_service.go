package service

import (
	"context"
	"fmt"
	"time"
)

func (s *AuthService) isRateLimited(ctx context.Context, email string) (bool, error) {
	key := fmt.Sprintf("login_attempt:%s", email)

	attempts, err := s.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if attempts == 1 {
		s.Redis.Expire(ctx, key, time.Minute)
	}

	return attempts > 5, nil
}
