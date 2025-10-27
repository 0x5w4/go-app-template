package app

import (
	"context"
	"errors"
	"fmt"
	"goapptemp/config"
	"goapptemp/internal/adapter/api/rest"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/adapter/repository"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared/token"
	"goapptemp/pkg/apmtracer"
	"goapptemp/pkg/bundb"
	"goapptemp/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pubsubClient "goapptemp/pkg/pubsub"
)

type App struct {
	config     *config.Config
	restServer rest.Server
	logger     logger.Logger
	tracer     apmtracer.Tracer
	pubsub     pubsubClient.Pubsub
}

func NewApp(config *config.Config, logger logger.Logger) (*App, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}

	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	return &App{
		config: config,
		logger: logger,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var (
		wg  sync.WaitGroup
		err error
	)

	// Initialize tracer
	a.tracer, err = apmtracer.NewApmTracer(&apmtracer.Config{
		ServiceName:    a.config.Tracer.ServiceName,
		ServiceVersion: a.config.Tracer.ServiceVersion,
		ServerURL:      a.config.Tracer.ServerURL,
		SecretToken:    a.config.Tracer.SecretToken,
		Environment:    a.config.Tracer.Environment,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Initialize repository
	repo, err := repository.NewRepository(a.config, a.logger)
	if err != nil {
		return fmt.Errorf("failed to setup repository: %w", err)
	}

	// Initialize pubsub
	var publisher pubsub.Publisher
	if a.config.App.UsePubsub {
		a.pubsub, err = pubsubClient.NewPubsub(ctx, a.config.Pubsub.ProjectID, a.config.Pubsub.CredFile)
		if err != nil {
			return fmt.Errorf("failed to setup pubsub client: %w", err)
		}

		publisher, err = pubsub.NewPublisher(a.logger, a.pubsub, a.config.Pubsub.TopicID)
		if err != nil {
			return fmt.Errorf("failed to setup pubsub publisher: %w", err)
		}
	}

	// Initialize token
	token, err := token.NewJwtToken(
		a.config.Token.AccessSecretKey,
		a.config.Token.RefreshSecretKey,
		time.Duration(a.config.Token.AccessTokenDuration)*time.Minute,
		time.Duration(a.config.Token.RefreshTokenDuration)*time.Minute,
	)
	if err != nil {
		return fmt.Errorf("failed to create token manager: %w", err)
	}

	// Initialize service
	service, err := service.NewService(a.config, repo, a.logger, token, publisher)
	if err != nil {
		return fmt.Errorf("failed to setup service: %w", err)
	}

	wg.Go(func() {
		service.StaleTaskDetector().Start(ctx)
	})

	// Initialize and start REST server
	a.restServer, err = rest.NewEchoServer(a.config, a.logger, token, service, repo)
	if err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	if err := a.restServer.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	a.logger.Info().Msgf("Server started at %s:%d", a.config.HTTP.Host, a.config.HTTP.Port)

	// Wait for shutdown signal
	<-ctx.Done()
	a.logger.Info().Msg("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Shutdown REST server
	if err := a.restServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error().Err(err).Msg("Failed to gracefully shutdown REST server")
	} else {
		a.logger.Info().Msg("REST server shut down gracefully")
	}

	// Wait for background tasks to finish
	a.logger.Info().Msg("Waiting for background tasks to finish...")
	wg.Wait()
	a.logger.Info().Msg("All background tasks finished")

	// Close repository
	if err := repo.Close(); err != nil {
		a.logger.Error().Err(err).Msg("Failed to gracefully close repository")
	} else {
		a.logger.Info().Msg("Repository closed gracefully")
	}

	// Shutdown pubsub client
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
	db, err := bundb.NewBunDB(a.config, a.logger)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			a.logger.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

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
