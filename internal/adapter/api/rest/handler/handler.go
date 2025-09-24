package handler

import (
	"errors"
	"fmt"
	"goapptemp/config"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/pkg/logger"

	validator "github.com/go-playground/validator/v10"
	"github.com/uptrace/bun"
)

type Handler interface {
	Auth() *AuthHandler
	City() *CityHandler
	District() *DistrictHandler
	Health() *HealthHandler
	Migration() *MigrationHandler
	Province() *ProvinceHandler
	Role() *RoleHandler
	SupportFeature() *SupportFeatureHandler
	User() *UserHandler
	Webhook() *WebhookHandler
}

type properties struct {
	config   *config.Config
	logger   logger.Logger
	service  service.Service
	validate *validator.Validate
}

type handler struct {
	properties
	authHandler           *AuthHandler
	cityHandler           *CityHandler
	districtHandler       *DistrictHandler
	healthHandler         *HealthHandler
	migrationHandler      *MigrationHandler
	provinceHandler       *ProvinceHandler
	roleHandler           *RoleHandler
	supportFeatureHandler *SupportFeatureHandler
	userHandler           *UserHandler
	webhookHandler        *WebhookHandler
}

func NewHandler(config *config.Config, logger logger.Logger, service service.Service, db *bun.DB) (*handler, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	validate, err := shared.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to setup validator: %w", err)
	}

	properties := properties{
		config:   config,
		service:  service,
		logger:   logger,
		validate: validate,
	}

	return &handler{
		properties:            properties,
		authHandler:           NewAuthHandler(properties),
		cityHandler:           NewCityHandler(properties),
		districtHandler:       NewDistrictHandler(properties),
		healthHandler:         NewHealthHandler(db, logger),
		migrationHandler:      NewMigrationHandler(properties),
		provinceHandler:       NewProvinceHandler(properties),
		roleHandler:           NewRoleHandler(properties),
		supportFeatureHandler: NewSupportFeatureHandler(properties),
		userHandler:           NewUserHandler(properties),
		webhookHandler:        NewWebhookHandler(properties),
	}, nil
}

func (h *handler) Auth() *AuthHandler {
	return h.authHandler
}

func (h *handler) City() *CityHandler {
	return h.cityHandler
}

func (h *handler) District() *DistrictHandler {
	return h.districtHandler
}

func (h *handler) Health() *HealthHandler {
	return h.healthHandler
}

func (h *handler) Migration() *MigrationHandler {
	return h.migrationHandler
}

func (h *handler) Province() *ProvinceHandler {
	return h.provinceHandler
}

func (h *handler) Role() *RoleHandler {
	return h.roleHandler
}

func (h *handler) SupportFeature() *SupportFeatureHandler {
	return h.supportFeatureHandler
}

func (h *handler) User() *UserHandler {
	return h.userHandler
}

func (h *handler) Webhook() *WebhookHandler {
	return h.webhookHandler
}
