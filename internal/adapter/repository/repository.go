package repository

import (
	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/repository/mysql"

	"go.elastic.co/apm/v2"
)

type Repository interface {
	MySQL() mysql.MySQLRepository
	Close() error
}

type repository struct {
	mysql mysql.MySQLRepository
}

func (r *repository) MySQL() mysql.MySQLRepository {
	return r.mysql
}

func NewRepository(config *config.Config, logger logger.Logger, tracer *apm.Tracer) (Repository, error) {
	mysqlRepo, err := mysql.NewMySQLRepository(config, logger, tracer)
	if err != nil {
		return nil, err
	}

	return &repository{
		mysql: mysqlRepo,
	}, nil
}

func (r *repository) Close() error {
	return r.mysql.Close()
}
