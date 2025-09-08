package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"goapptemp/config"
	restServer "goapptemp/internal/adapter/api/rest/server"
	"goapptemp/internal/adapter/elastic/tracer"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/repository/mysql/db"
	"goapptemp/internal/domain/service"
)

type App struct {
	cfg        *config.Config
	restServer restServer.Server
	logger     logger.Logger
	tracer     tracer.Tracer
	service    service.Service
	pubsub     pubsub.Pubsub
}

func NewApp(cfg *config.Config, logger logger.Logger) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &App{
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg sync.WaitGroup

	var err error

	a.tracer, err = tracer.InitTracer(&tracer.Config{
		ServiceName:    a.cfg.Tracer.ServiceName,
		ServiceVersion: a.cfg.Tracer.ServiceVersion,
		ServerURL:      a.cfg.Tracer.ServerURL,
		SecretToken:    a.cfg.Tracer.SecretToken,
		Environment:    a.cfg.Tracer.Environment,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	repo, err := repository.NewRepository(a.cfg, a.logger, a.tracer.Tracer())
	if err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	if a.cfg.App.UsePubsub {
		a.pubsub, err = pubsub.NewPubSubClient(ctx, a.cfg.Pubsub.ProjectID, a.cfg.Pubsub.CredFile, a.cfg.Pubsub.TopicID)
		if err != nil {
			return fmt.Errorf("failed to setup pubsub: %w", err)
		}
	}

	a.service, err = service.NewService(a.cfg, repo, a.logger, a.pubsub)
	if err != nil {
		return fmt.Errorf("failed to setup service: %w", err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		a.service.StaleTaskDetector().Start(ctx)
	}()

	a.restServer, err = restServer.NewServer(a.cfg, a.logger, a.service, repo, a.tracer)
	if err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	if err := a.restServer.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	<-ctx.Done()
	a.logger.Info().Msg("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := a.restServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error().Err(err).Msg("Failed to gracefully shutdown REST server")
	} else {
		a.logger.Info().Msg("REST server shut down gracefully")
	}

	a.logger.Info().Msg("Waiting for background tasks to finish...")
	wg.Wait()
	a.logger.Info().Msg("All background tasks finished")

	if err := repo.Close(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to gracefully close repository")
	} else {
		a.logger.Info().Msg("Repository closed gracefully")
	}

	if a.pubsub != nil {
		if err := a.pubsub.Shutdown(); err != nil {
			a.logger.Error().Err(err).Msg("Failed to gracefully shutdown PubSub client")
		} else {
			a.logger.Info().Msg("PubSub client shut down gracefully")
		}
	}

	a.tracer.Shutdown()

	return nil
}

func (a *App) Migrate(reset bool) error {
	db, err := db.NewBunDB(a.cfg, a.logger, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if reset {
		if err := db.Reset(); err != nil {
			return err
		}
	} else {
		if err := db.Migrate(); err != nil {
			return err
		}
	}

	return nil
}
