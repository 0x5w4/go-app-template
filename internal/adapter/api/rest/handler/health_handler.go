package handler

import (
	"context"
	"goapptemp/pkg/logger"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type HealthHandler struct {
	db     *bun.DB
	logger logger.Logger
}

func NewHealthHandler(db *bun.DB, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

func (h *HealthHandler) CheckHealth(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error().Err(err).Msg("Health check failed: database connection is down")

		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"status":   "error",
			"database": "unhealthy",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":   "ok",
		"database": "healthy",
	})
}
