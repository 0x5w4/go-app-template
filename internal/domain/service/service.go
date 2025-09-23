package service

import (
	"goapptemp/config"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/adapter/repository"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/token"
	"goapptemp/pkg/logger"
)

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
	Webhook() WebhookService
	StaleTaskDetector() StaleTaskDetector
}

type services struct {
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
}

func NewService(
	config *config.Config,
	repo repository.Repository,
	logger logger.Logger,
	token token.Token,
	publisher pubsub.Publisher,
) (Service, error) {
	validate, err := shared.NewValidator()
	if err != nil {
		return nil, err
	}

	pubsubService := NewPubsubService(config, logger, publisher)
	authService := NewAuthService(config, token, repo, logger)

	return &services{
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
	}, nil
}

func (s *services) Token() token.Token {
	return s.tokenManager
}

func (s *services) Auth() AuthService {
	return s.authService
}

func (s *services) User() UserService {
	return s.userService
}

func (s *services) Client() ClientService {
	return s.clientService
}

func (s *services) Role() RoleService {
	return s.roleService
}

func (s *services) SupportFeature() SupportFeatureService {
	return s.supportFeatureService
}

func (s *services) Province() ProvinceService {
	return s.provinceService
}

func (s *services) City() CityService {
	return s.cityService
}

func (s *services) District() DistrictService {
	return s.districtService
}

func (s *services) Webhook() WebhookService {
	return s.webhookService
}

func (s *services) StaleTaskDetector() StaleTaskDetector {
	return s.staleTaskDetector
}
