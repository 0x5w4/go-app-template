package redisrepository

import (
	"context"
	"goapptemp/config"
	"goapptemp/pkg/logger"
	redisclient "goapptemp/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ RedisRepository = (*redisRepository)(nil)

type RedisRepository interface {
	Close() error
	CheckLockedUserExists(ctx context.Context, phone string) (bool, error)
	GetBlockIPTTL(ctx context.Context, ip string) (time.Duration, error)
	RecordUserFailure(ctx context.Context, phone string) error
	RecordIPFailure(ctx context.Context, ip string) (blockNow bool, retryAfter int, err error)
	DeleteUserAttempts(ctx context.Context, identifier string) error
	DeleteIPAttempts(ctx context.Context, ip string) error
	DeleteBlockCount(ctx context.Context, ip string) error
	BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error
	CheckTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	StoreResetToken(ctx context.Context, token string, userID uint, ttl time.Duration) error
	GetUserIDFromResetToken(ctx context.Context, token string) (uint, error)
	DeleteResetToken(ctx context.Context, token string) error
}

type redisRepository struct {
	db *redis.Client
}

func NewRedisRepository(config *config.Config, logger logger.Logger) (*redisRepository, error) {
	db, err := redisclient.NewRedisClient(config, logger)
	if err != nil {
		return nil, err
	}

	return &redisRepository{
		db: db.DB(),
	}, nil
}

func (r *redisRepository) Close() error {
	return r.db.Close()
}
