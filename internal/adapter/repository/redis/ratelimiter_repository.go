package redisrepository

import (
	"context"
	"fmt"
	"goapptemp/constant"
	"math"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

func (r *redisRepository) CheckLockedUserExists(ctx context.Context, identifier string) (bool, error) {
	userLockKey := fmt.Sprintf(KeyPatternUserLock, identifier)

	count, err := r.db.Exists(ctx, userLockKey).Result()
	if err != nil {
		return false, handleRedisError(err, "check if locked user exists")
	}

	return count > 0, nil
}

func (r *redisRepository) GetBlockIPTTL(ctx context.Context, ip string) (time.Duration, error) {
	ipBlockKey := fmt.Sprintf(KeyPatternBlockIP, ip)

	ttl, err := r.db.TTL(ctx, ipBlockKey).Result()
	if err != nil {
		return 0, handleRedisError(err, "get block IP TTL")
	}

	return ttl, nil
}

func (r *redisRepository) RecordUserFailure(ctx context.Context, identifier string) error {
	userAttemptKey := fmt.Sprintf(KeyPatternUserAttempts, identifier)

	failCount, err := r.db.Incr(ctx, userAttemptKey).Result()
	if err != nil {
		return handleRedisError(err, "increment failed user attempts")
	}

	if failCount == 1 {
		r.db.Expire(ctx, userAttemptKey, constant.UserFailedWindow)
	}

	if failCount >= int64(constant.UserFailedAttemptsLimit) {
		lockKey := fmt.Sprintf(KeyPatternUserLock, identifier)

		r.db.Set(ctx, lockKey, "1", constant.UserLockoutDuration)
		r.db.Del(ctx, userAttemptKey)
	}

	return nil
}

func (r *redisRepository) RecordIPFailure(ctx context.Context, ip string) (blockNow bool, retryAfter int, err error) {
	ipAttemptKey := fmt.Sprintf(KeyPatternIPAttempts, ip)

	failCount, err := r.db.Incr(ctx, ipAttemptKey).Result()
	if err != nil {
		return false, 0, handleRedisError(err, "increment failed IP attempts")
	}

	if failCount == 1 {
		r.db.Expire(ctx, ipAttemptKey, constant.IpRateLimitWindow)
	}

	if failCount >= int64(constant.IpRateLimitAttempts) {
		blockCountKey := fmt.Sprintf(KeyPatternBlockCountIP, ip)

		blockLevel, err := r.db.Incr(ctx, blockCountKey).Result()
		if err != nil {
			return false, 0, handleRedisError(err, "increment block count for IP")
		}

		durationSeconds := int(float64(constant.IpBackoffBaseSeconds) * math.Pow(2, float64(blockLevel-1)))
		blockDuration := time.Duration(durationSeconds) * time.Second

		ipBlockKey := fmt.Sprintf(KeyPatternBlockIP, ip)
		r.db.Set(ctx, ipBlockKey, "1", blockDuration)
		r.db.Del(ctx, ipAttemptKey)

		return true, durationSeconds, nil
	}

	return false, 0, nil
}

func (r *redisRepository) DeleteUserAttempts(ctx context.Context, identifier string) error {
	userAttemptKey := fmt.Sprintf(KeyPatternUserAttempts, identifier)
	err := r.db.Del(ctx, userAttemptKey).Err()

	return handleRedisError(err, "delete user attempts")
}

func (r *redisRepository) DeleteIPAttempts(ctx context.Context, ip string) error {
	ipAttemptsKey := fmt.Sprintf(KeyPatternIPAttempts, ip)
	err := r.db.Del(ctx, ipAttemptsKey).Err()

	return handleRedisError(err, "delete IP attempts")
}

func (r *redisRepository) DeleteBlockCount(ctx context.Context, ip string) error {
	blockCountIPKey := fmt.Sprintf(KeyPatternBlockCountIP, ip)
	err := r.db.Del(ctx, blockCountIPKey).Err()

	return handleRedisError(err, "delete block count")
}

func (r *redisRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	blacklistTokenKey := fmt.Sprintf(KeyPatternBlacklistToken, jti)
	err := r.db.Set(ctx, blacklistTokenKey, "1", ttl).Err()

	return handleRedisError(err, "blacklist token")
}

func (r *redisRepository) CheckTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	blacklistTokenKey := fmt.Sprintf(KeyPatternBlacklistToken, jti)
	err := r.db.Get(ctx, blacklistTokenKey).Err()

	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, handleRedisError(err, "check if token is blacklisted")
	}

	return true, nil
}
