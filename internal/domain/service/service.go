package service

import (
	"time"

	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/pubsub"
	repo "goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/util"
	"goapptemp/internal/adapter/util/token"
	"goapptemp/internal/domain/service/auth"
	"goapptemp/internal/domain/service/city"
	"goapptemp/internal/domain/service/client"
	"goapptemp/internal/domain/service/district"
	"goapptemp/internal/domain/service/province"
	pubsubService "goapptemp/internal/domain/service/pubsub"
	"goapptemp/internal/domain/service/role"
	"goapptemp/internal/domain/service/staletask"
	"goapptemp/internal/domain/service/supportfeature"
	"goapptemp/internal/domain/service/user"
	"goapptemp/internal/domain/service/webhook"
)

type Service interface {
	Token() token.Token
	Auth() auth.AuthService
	User() user.UserService
	Client() client.ClientService
	Role() role.RoleService
	SupportFeature() supportfeature.SupportFeatureService
	Province() province.ProvinceService
	City() city.CityService
	District() district.DistrictService
	Webhook() webhook.WebhookService
	StaleTaskDetector() staletask.StaleTaskDetector
}

type services struct {
	tokenManager          token.Token
	authService           auth.AuthService
	userService           user.UserService
	clientService         client.ClientService
	roleService           role.RoleService
	supportFeatureService supportfeature.SupportFeatureService
	webhookService        webhook.WebhookService
	provinceService       province.ProvinceService
	cityService           city.CityService
	districtService       district.DistrictService
	staleTaskDetector     staletask.StaleTaskDetector
}

func NewService(cfg *config.Config, repo repo.Repository, logger logger.Logger, pubsub pubsub.Pubsub) (Service, error) {
	tokenManager, err := token.NewJWTManager(cfg.Token.SecretKey, time.Duration(cfg.Token.ExpireTime)*time.Second)
	if err != nil {
		return nil, err
	}

	validate, err := util.SetupValidator()
	if err != nil {
		return nil, err
	}

	pubsubService := pubsubService.NewPubsubService(cfg, logger, pubsub)
	authService := auth.NewAuthService(cfg, tokenManager, repo, logger)

	return &services{
		tokenManager:          tokenManager,
		authService:           authService,
		userService:           user.NewUserService(cfg, repo, logger, authService),
		clientService:         client.NewClientService(cfg, repo, logger, authService, pubsubService),
		roleService:           role.NewRoleService(cfg, repo, logger, authService),
		supportFeatureService: supportfeature.NewSupportFeatureService(cfg, repo, logger, authService, validate),
		provinceService:       province.NewProvinceService(cfg, repo, logger, authService),
		cityService:           city.NewCityService(cfg, repo, logger, authService),
		districtService:       district.NewDistrictService(cfg, repo, logger, authService),
		staleTaskDetector:     staletask.NewStaleTaskDetector(cfg, repo, logger),
		webhookService:        webhook.NewWebhookService(cfg, repo, logger),
	}, nil
}

func (s *services) Token() token.Token {
	return s.tokenManager
}

func (s *services) Auth() auth.AuthService {
	return s.authService
}

func (s *services) User() user.UserService {
	return s.userService
}

func (s *services) Client() client.ClientService {
	return s.clientService
}

func (s *services) Role() role.RoleService {
	return s.roleService
}

func (s *services) SupportFeature() supportfeature.SupportFeatureService {
	return s.supportFeatureService
}

func (s *services) Province() province.ProvinceService {
	return s.provinceService
}

func (s *services) City() city.CityService {
	return s.cityService
}

func (s *services) District() district.DistrictService {
	return s.districtService
}

func (s *services) Webhook() webhook.WebhookService {
	return s.webhookService
}

func (s *services) StaleTaskDetector() staletask.StaleTaskDetector {
	return s.staleTaskDetector
}
