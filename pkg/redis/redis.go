package redisclient

import (
	"context"
	"fmt"
	"goapptemp/config"
	"goapptemp/pkg/logger"
	redishook "goapptemp/pkg/redis/hook"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ RedisClient = (*redisClient)(nil)

type RedisClient interface {
	DB() *redis.Client
	Close() error
}

type redisClient struct {
	config *config.Config
	logger logger.Logger
	db     *redis.Client
}

func NewRedisClient(config *config.Config, logger logger.Logger) (*redisClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	slowThreshold := 100 * time.Millisecond

	tracerHook := redishook.NewTracerHook()
	loggerHook := redishook.NewLoggerHook(
		redishook.WithLogger(logger),
		redishook.WithSlowThreshold(slowThreshold),
		redishook.WithDebug(config.App.Debug),
	)

	db.AddHook(tracerHook)
	db.AddHook(loggerHook)

	if _, err := db.Ping(ctx).Result(); err != nil {
		log.Fatalf("Tidak dapat terhubung ke Redis: %v", err)
	}

	return &redisClient{
		config: config,
		logger: logger,
		db:     db,
	}, nil
}

func (r *redisClient) DB() *redis.Client {
	return r.db
}

func (r *redisClient) Close() error {
	if r.db != nil {
		return r.db.Close()
	}

	return nil
}
