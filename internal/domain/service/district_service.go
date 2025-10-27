package service

import (
	"context"
	"goapptemp/config"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	repo "goapptemp/internal/adapter/repository"

	serror "goapptemp/internal/domain/service/error"
)

var _ DistrictService = (*districtService)(nil)

type DistrictService interface {
	Find(ctx context.Context, req *FindDistrictsRequest) ([]*entity.District, int, error)
	FindOne(ctx context.Context, req *FindOneDistrictRequest) (*entity.District, error)
}

type districtService struct {
	config *config.Config
	repo   repo.Repository
	log    logger.Logger
	auth   AuthService
}

func NewDistrictService(config *config.Config, repo repo.Repository, log logger.Logger, auth AuthService) *districtService {
	return &districtService{
		config: config,
		repo:   repo,
		log:    log,
		auth:   auth,
	}
}

type FindDistrictsRequest struct {
	Filter *mysqlrepository.FilterDistrictPayload
}

func (s *districtService) Find(ctx context.Context, req *FindDistrictsRequest) ([]*entity.District, int, error) {
	districts, totalCount, err := s.repo.MySQL().District().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return districts, totalCount, nil
}

type FindOneDistrictRequest struct {
	DistrictID uint
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
