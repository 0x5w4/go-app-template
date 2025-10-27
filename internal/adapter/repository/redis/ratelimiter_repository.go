package redisrepository

import (
	"context"
	"fmt"
	"goapptemp/constant"
	"math"
	"time"
)

func (r *redisRepository) CheckLockedUserExists(ctx context.Context, identifier string) (bool, error) {
	userLockKey := fmt.Sprintf("lock:user:%s", identifier)
	count, err := r.db.Exists(ctx, userLockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock user exist: %w", err)
	}
	return count > 0, nil
}

func (r *redisRepository) GetBlockIPTTL(ctx context.Context, ip string) (time.Duration, error) {
	ipBlockKey := fmt.Sprintf("block:ip:%s", ip)
	ttl, err := r.db.TTL(ctx, ipBlockKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get block ip ttl: %w", err)
	}
	return ttl, nil
}

func (r *redisRepository) RecordUserFailure(ctx context.Context, identifier string) error {
	userAttemptKey := fmt.Sprintf("attempts:user:%s", identifier)
	failCount, err := r.db.Incr(ctx, userAttemptKey).Result()
	if err != nil {
		return fmt.Errorf("failed to increment failed user attempts: %w", err)
	}

	if failCount == 1 {
		r.db.Expire(ctx, userAttemptKey, constant.UserFailedWindow)
	}

	if failCount >= int64(constant.UserFailedAttemptsLimit) {
		lockKey := fmt.Sprintf("lock:user:%s", identifier)
		r.db.Set(ctx, lockKey, "locked", constant.UserLockoutDuration)
		r.db.Del(ctx, userAttemptKey)
	}

	return nil
}

func (r *redisRepository) RecordIPFailure(ctx context.Context, ip string) (blockNow bool, retryAfter int, err error) {
	ipAttemptKey := fmt.Sprintf("attempts:ip:%s", ip)
	failCount, err := r.db.Incr(ctx, ipAttemptKey).Result()
	if err != nil {
		return false, 0, fmt.Errorf("failed to increment failed ip attempts: %w", err)
	}

	if failCount == 1 {
		r.db.Expire(ctx, ipAttemptKey, constant.IpRateLimitWindow)
	}

	if failCount >= int64(constant.IpRateLimitAttempts) {
		blockCountKey := fmt.Sprintf("blockcount:ip:%s", ip)
		blockLevel, err := r.db.Incr(ctx, blockCountKey).Result()
		if err != nil {
			return false, 0, fmt.Errorf("failed to increment block count for ip: %w", err)
		}

		durationSeconds := int(float64(constant.IpBackoffBaseSeconds) * math.Pow(2, float64(blockLevel-1)))
		blockDuration := time.Duration(durationSeconds) * time.Second

		ipBlockKey := fmt.Sprintf("block:ip:%s", ip)
		r.db.Set(ctx, ipBlockKey, "blocked", blockDuration)
		r.db.Del(ctx, ipAttemptKey)

		return true, durationSeconds, nil
	}

	return false, 0, nil
}

func (r *redisRepository) DeleteUserAttempt(ctx context.Context, identifier string) error {
	err := r.db.Del(ctx, fmt.Sprintf("attempts:user:%s", identifier)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user attempts: %w", err)
	}
	return nil
}

func (r *redisRepository) DeleteIPAttempt(ctx context.Context, ip string) error {
	err := r.db.Del(ctx, fmt.Sprintf("attempts:ip:%s", ip)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete ip attempts: %w", err)
	}
	return nil
}

func (r *redisRepository) DeleteBlockCount(ctx context.Context, ip string) error {
	err := r.db.Del(ctx, fmt.Sprintf("blockcount:ip:%s", ip)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}
