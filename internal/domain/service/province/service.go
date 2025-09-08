package province

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

type ProvinceService interface {
	Find(ctx context.Context, req *FindProvincesRequest) ([]*entity.Province, int, error)
	FindOne(ctx context.Context, req *FindOneProvinceRequest) (*entity.Province, error)
}

type provinceService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
	auth   auth.AuthService
}

func NewProvinceService(config *config.Config, repo repo.Repository, logger logger.Logger, auth auth.AuthService) ProvinceService {
	return &provinceService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

func (s *provinceService) Find(ctx context.Context, req *FindProvincesRequest) ([]*entity.Province, int, error) {
	provinces, totalCount, err := s.repo.MySQL().Province().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return provinces, totalCount, nil
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
