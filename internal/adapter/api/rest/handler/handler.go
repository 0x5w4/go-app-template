package handler

import (
	"errors"
	"fmt"
	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/util"
	"goapptemp/internal/adapter/util/exception"
	"goapptemp/internal/adapter/util/token"
	"goapptemp/internal/domain/service"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	config   *config.Config
	logger   logger.Logger
	service  service.Service
	validate *validator.Validate
}

func NewHandler(config *config.Config, logger logger.Logger, service service.Service) (*Handler, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	v, err := util.SetupValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to setup validator: %w", err)
	}

	return &Handler{
		config:   config,
		service:  service,
		logger:   logger,
		validate: v,
	}, nil
}

func (h *Handler) VerifyToken(tokenStr string) (*token.Claims, error) {
	claims, err := h.service.Token().Verify(tokenStr)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeTokenInvalid, exception.CodeTokenInvalid, "Token verification failed")
	}

	return claims, nil
}
