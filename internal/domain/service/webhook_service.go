package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/constant"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/pkg/logger"
	"strings"

	repo "goapptemp/internal/adapter/repository"

	serror "goapptemp/internal/domain/service/error"
)

var _ WebhookService = (*webhookService)(nil)

type WebhookService interface {
	UpdateIcon(ctx context.Context, req *UpdateIconRequest) error
}

type webhookService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
}

func NewWebhookService(config *config.Config, repo repo.Repository, logger logger.Logger) *webhookService {
	return &webhookService{
		config: config,
		repo:   repo,
		logger: logger,
	}
}

type UpdateIconRequest struct {
	ID   uint
	Type string
	Link string
}

func (s *webhookService) UpdateIcon(ctx context.Context, req *UpdateIconRequest) error {
	if req.Type == "client" {
		client, err := s.repo.MySQL().Client().FindByID(ctx, req.ID, false)
		if err != nil {
			return serror.TranslateRepoError(err)
		}

		if *client.Icon == constant.FailedIcon || strings.Contains(*client.Icon, "http://") || strings.Contains(*client.Icon, "https://") {
			return nil
		}

		_, err = s.repo.MySQL().Client().Update(ctx, &mysqlrepository.UpdateClientPayload{
			ID:   req.ID,
			Icon: &req.Link,
		})
		if err != nil {
			return serror.TranslateRepoError(err)
		}
	}

	return nil
}
