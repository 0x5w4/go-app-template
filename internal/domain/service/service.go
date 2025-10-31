package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/adapter/repository"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/token"
	"goapptemp/pkg/gmailsender"
	"goapptemp/pkg/logger"
)

var _ Service = (*service)(nil)

type Service interface {
	Token() token.Token
	Auth() AuthService
	User() UserService
	Client() ClientService
	Role() RoleService
	SupportFeature() SupportFeatureService
	Province() ProvinceService
	City() CityService
	District() DistrictService
	Notification() NotificationService
	Webhook() WebhookService
	StaleTaskDetector() StaleTaskDetector
}

type service struct {
	tokenManager          token.Token
	authService           AuthService
	userService           UserService
	clientService         ClientService
	roleService           RoleService
	supportFeatureService SupportFeatureService
	webhookService        WebhookService
	provinceService       ProvinceService
	cityService           CityService
	districtService       DistrictService
	staleTaskDetector     StaleTaskDetector
	notificationService   NotificationService
}

func NewService(
	config *config.Config,
	repo repository.Repository,
	logger logger.Logger,
	token token.Token,
	publisher pubsub.Publisher,
) (*service, error) {
	validate, err := shared.NewValidator()
	if err != nil {
		return nil, err
	}

	gmailSender, err := gmailsender.NewGmailSender(context.Background(), config.Gmail.CredFile, config.Gmail.CredFile, config.Gmail.Sender)
	if err != nil {
		return nil, err
	}

	notifService := NewNotificationService(gmailSender, logger)
	pubsubService := NewPubsubService(config, logger, publisher)
	authService := NewAuthService(config, token, repo, logger, notifService)

	return &service{
		authService:           authService,
		userService:           NewUserService(config, repo, logger, authService),
		clientService:         NewClientService(config, repo, logger, authService, pubsubService),
		roleService:           NewRoleService(config, repo, logger, authService),
		supportFeatureService: NewSupportFeatureService(config, repo, logger, authService, validate),
		provinceService:       NewProvinceService(config, repo, logger, authService),
		cityService:           NewCityService(config, repo, logger, authService),
		districtService:       NewDistrictService(config, repo, logger, authService),
		staleTaskDetector:     NewStaleTaskDetector(config, repo, logger),
		webhookService:        NewWebhookService(config, repo, logger),
		notificationService:   notifService,
	}, nil
}

func (s *service) Token() token.Token {
	return s.tokenManager
}

func (s *service) Auth() AuthService {
	return s.authService
}

func (s *service) User() UserService {
	return s.userService
}

func (s *service) Client() ClientService {
	return s.clientService
}

func (s *service) Role() RoleService {
	return s.roleService
}

func (s *service) SupportFeature() SupportFeatureService {
	return s.supportFeatureService
}

func (s *service) Province() ProvinceService {
	return s.provinceService
}

func (s *service) City() CityService {
	return s.cityService
}

func (s *service) District() DistrictService {
	return s.districtService
}

func (s *service) Notification() NotificationService {
	return s.notificationService
}

func (s *service) Webhook() WebhookService {
	return s.webhookService
}

func (s *service) StaleTaskDetector() StaleTaskDetector {
	return s.staleTaskDetector
}
