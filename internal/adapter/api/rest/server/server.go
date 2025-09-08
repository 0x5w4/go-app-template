package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"goapptemp/config"
	"goapptemp/internal/adapter/api/rest/handler"
	"goapptemp/internal/adapter/elastic/tracer"
	"goapptemp/internal/adapter/logger"
	repo "goapptemp/internal/adapter/repository"
	"goapptemp/internal/domain/service"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
)

const shutdownTimeout = 10 * time.Second

type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type server struct {
	config  *config.Config
	logger  logger.Logger
	tracer  tracer.Tracer
	handler *handler.Handler
	echo    *echo.Echo
	repo    repo.Repository
}

func NewServer(cfg *config.Config, logger logger.Logger, service service.Service, repo repo.Repository, tracer tracer.Tracer) (Server, error) {
	e := echo.New()
	e.HideBanner = true

	h, err := handler.NewHandler(cfg, logger, service)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create handler")
	}

	server := &server{
		config:  cfg,
		logger:  logger.NewInstance().Field("component", "http_server").Logger(),
		handler: h,
		echo:    e,
		repo:    repo,
		tracer:  tracer,
	}
	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

func (s *server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	s.logger.Info().Msgf("Starting server on %s", address)

	startErrChan := make(chan error, 1)

	go func() {
		if err := s.echo.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("Server failed to run")
			startErrChan <- errors.Wrapf(err, "server failed to start listening on %s", address)
		} else {
			startErrChan <- nil
		}

		close(startErrChan)
	}()
	select {
	case err := <-startErrChan:
		if err != nil {
			return err
		}

		s.logger.Info().Msg("Server listening")
	case <-time.After(100 * time.Millisecond):
		s.logger.Info().Msg("Server assumed started successfully (listening)")
	}

	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.logger.Error().Err(err).Msg("Server shutdown failed")
		return errors.Wrap(err, "server shutdown failed")
	}

	return nil
}
