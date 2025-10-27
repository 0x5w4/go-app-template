package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/internal/adapter/repository"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	serror "goapptemp/internal/domain/service/error"
)

var _ CityService = (*cityService)(nil)

type CityService interface {
	Find(ctx context.Context, req *FindCitiesRequest) ([]*entity.City, int, error)
	FindOne(ctx context.Context, req *FindOneCityRequest) (*entity.City, error)
}

type cityService struct {
	config *config.Config
	repo   repository.Repository
	log    logger.Logger
	auth   AuthService
}

func NewCityService(config *config.Config, repo repository.Repository, log logger.Logger, auth AuthService) *cityService {
	return &cityService{
		config: config,
		repo:   repo,
		log:    log,
		auth:   auth,
	}
}

type FindCitiesRequest struct {
	Filter *mysqlrepository.FilterCityPayload
}

func (s *cityService) Find(ctx context.Context, req *FindCitiesRequest) ([]*entity.City, int, error) {
	citys, totalCount, err := s.repo.MySQL().City().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return citys, totalCount, nil
}

type FindOneCityRequest struct {
	CityID uint
}

func (s *cityService) FindOne(ctx context.Context, req *FindOneCityRequest) (*entity.City, error) {
	if req.CityID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "City ID required for find one")
	}

	city, err := s.repo.MySQL().City().FindByID(ctx, req.CityID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return city, nil
}
