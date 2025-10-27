package repository

import (
	"goapptemp/config"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	redisrepository "goapptemp/internal/adapter/repository/redis"
	"goapptemp/pkg/logger"
)

type Repository interface {
	MySQL() mysqlrepository.MySQLRepository
	Redis() redisrepository.RedisRepository
	Close() error
}

type repository struct {
	mysql mysqlrepository.MySQLRepository
	redis redisrepository.RedisRepository
}

func NewRepository(config *config.Config, logger logger.Logger) (Repository, error) {
	mysqlRepo, err := mysqlrepository.NewMySQLRepository(config, logger)
	if err != nil {
		return nil, err
	}

	redisRepo, err := redisrepository.NewRedisRepository(config, logger)
	if err != nil {
		return nil, err
	}

	return &repository{
		mysql: mysqlRepo,
		redis: redisRepo,
	}, nil
}

func (r *repository) MySQL() mysqlrepository.MySQLRepository {
	return r.mysql
}

func (r *repository) Redis() redisrepository.RedisRepository {
	return r.redis
}

func (r *repository) Close() error {
	return r.mysql.Close()
}
