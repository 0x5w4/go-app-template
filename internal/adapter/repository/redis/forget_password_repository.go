package redisrepository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
)

func (r *redisRepository) StoreResetToken(ctx context.Context, token string, userID uint, ttl time.Duration) error {
	key := fmt.Sprintf(KeyPatternResetPassword, token)

	err := r.db.Set(ctx, key, userID, ttl).Err()

	return handleRedisError(err, "store reset token")
}

func (r *redisRepository) GetUserIDFromResetToken(ctx context.Context, token string) (uint, error) {
	key := fmt.Sprintf(KeyPatternResetPassword, token)

	userIDStr, err := r.db.Get(ctx, key).Result()
	if err != nil {
		return 0, handleRedisError(err, "get reset token")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse user ID from redis")
	}

	return uint(userID), nil
}

func (r *redisRepository) DeleteResetToken(ctx context.Context, token string) error {
	key := fmt.Sprintf(KeyPatternResetPassword, token)
	err := r.db.Del(ctx, key).Err()

	return handleRedisError(err, "delete reset token")
}
