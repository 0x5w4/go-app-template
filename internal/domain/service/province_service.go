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

var _ ProvinceService = (*provinceService)(nil)

type ProvinceService interface {
	Find(ctx context.Context, req *FindProvincesRequest) ([]*entity.Province, int, error)
	FindOne(ctx context.Context, req *FindOneProvinceRequest) (*entity.Province, error)
}

type provinceService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
	auth   AuthService
}

func NewProvinceService(config *config.Config, repo repo.Repository, logger logger.Logger, auth AuthService) *provinceService {
	return &provinceService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

type FindProvincesRequest struct {
	Filter *mysqlrepository.FilterProvincePayload
}

func (s *provinceService) Find(ctx context.Context, req *FindProvincesRequest) ([]*entity.Province, int, error) {
	provinces, totalCount, err := s.repo.MySQL().Province().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return provinces, totalCount, nil
}

type FindOneProvinceRequest struct {
	ProvinceID uint
}

func (s *provinceService) FindOne(ctx context.Context, req *FindOneProvinceRequest) (*entity.Province, error) {
	if req.ProvinceID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Province ID required for find one")
	}

	province, err := s.repo.MySQL().Province().FindByID(ctx, req.ProvinceID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return province, nil
}
