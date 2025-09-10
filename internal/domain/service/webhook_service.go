package service

import (
	"context"
	"strings"

	"goapptemp/config"
	repo "goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/repository/mysql"
	serror "goapptemp/internal/domain/service/error"
	"goapptemp/pkg/logger"
)

type WebhookService interface {
	UpdateIcon(ctx context.Context, req *UpdateIconRequest) error
}

const (
	failedIcon = "failed"
)

type webhookService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
}

func NewWebhookService(config *config.Config, repo repo.Repository, logger logger.Logger) WebhookService {
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
	switch req.Type {
	case "client":
		client, err := s.repo.MySQL().Client().FindByID(ctx, req.ID, false)
		if err != nil {
			return serror.TranslateRepoError(err)
		}

		if *client.Icon == failedIcon || strings.Contains(*client.Icon, "http://") || strings.Contains(*client.Icon, "https://") {
			return nil
		}

		_, err = s.repo.MySQL().Client().Update(ctx, &mysql.UpdateClientPayload{
			ID:   req.ID,
			Icon: &req.Link,
		})
		if err != nil {
			return serror.TranslateRepoError(err)
		}
	}
	return nil
}
