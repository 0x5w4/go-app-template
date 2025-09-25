package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/pkg/logger"
	"time"

	repo "goapptemp/internal/adapter/repository"

	apm "go.elastic.co/apm/v2"
)

var _ StaleTaskDetector = (*staleTaskDetector)(nil)

type StaleTaskDetector interface {
	Start(ctx context.Context)
}

type staleTaskDetector struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
}

func NewStaleTaskDetector(config *config.Config, repo repo.Repository, logger logger.Logger) *staleTaskDetector {
	return &staleTaskDetector{
		config: config,
		repo:   repo,
		logger: logger,
	}
}

func (d *staleTaskDetector) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(d.config.StaleTask.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.checkAndProcessStaleTasks(ctx)
		case <-ctx.Done():
			d.logger.Info().Msg("Stale task detector stopping due to context cancellation.")
			return
		}
	}
}

func (d *staleTaskDetector) checkAndProcessStaleTasks(ctx context.Context) {
	span, ctx := apm.StartSpan(ctx, "StaleTaskDetector.checkAndProcessStaleTasks", "task")
	defer span.End()

	err := d.repo.MySQL().Client().UpdateStaleIcons(ctx)
	if err != nil {
		if apmErr := apm.CaptureError(ctx, err); apmErr != nil {
			apmErr.Handled = true
			apmErr.Send()
		}

		d.logger.Error().Err(err).Msg("Failed to update stale icons")
	}

	d.logger.Info().Msg("Stale task check completed")
}
