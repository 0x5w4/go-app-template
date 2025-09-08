package district

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

type DistrictService interface {
	Find(ctx context.Context, req *FindDistrictsRequest) ([]*entity.District, int, error)
	FindOne(ctx context.Context, req *FindOneDistrictRequest) (*entity.District, error)
}

type districtService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
	auth   auth.AuthService
}

func NewDistrictService(config *config.Config, repo repo.Repository, logger logger.Logger, auth auth.AuthService) DistrictService {
	return &districtService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

func (s *districtService) Find(ctx context.Context, req *FindDistrictsRequest) ([]*entity.District, int, error) {
	districts, totalCount, err := s.repo.MySQL().District().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return districts, totalCount, nil
}

func (s *districtService) FindOne(ctx context.Context, req *FindOneDistrictRequest) (*entity.District, error) {
	if req.DistrictID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "District ID required for find one")
	}

	district, err := s.repo.MySQL().District().FindByID(ctx, req.DistrictID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return district, nil
}
