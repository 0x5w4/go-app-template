package city

import (
	"context"

	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	repo "goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/util/exception"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/auth"
	serror "goapptemp/internal/domain/service/error"
)

type CityService interface {
	Find(ctx context.Context, req *FindCitiesRequest) ([]*entity.City, int, error)
	FindOne(ctx context.Context, req *FindOneCityRequest) (*entity.City, error)
}

type cityService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
	auth   auth.AuthService
}

func NewCityService(config *config.Config, repo repo.Repository, logger logger.Logger, auth auth.AuthService) CityService {
	return &cityService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

func (s *cityService) Find(ctx context.Context, req *FindCitiesRequest) ([]*entity.City, int, error) {
	citys, totalCount, err := s.repo.MySQL().City().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return citys, totalCount, nil
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
