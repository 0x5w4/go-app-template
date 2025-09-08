package logger

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/util/constant"

	"github.com/uptrace/bun"
)

var _ bun.QueryHook = (*QueryHook)(nil)

type Option func(h *QueryHook)

func WithLogger(logger logger.Logger) Option {
	return func(h *QueryHook) {
		h.logger = logger
	}
}

func WithDebug(debug bool) Option {
	return func(h *QueryHook) {
		h.debug = debug
	}
}

func WithSlowQueryThreshold(threshold time.Duration) Option {
	return func(h *QueryHook) {
		h.slowQueryThreshold = threshold
	}
}

type QueryHook struct {
	logger             logger.Logger
	debug              bool
	slowQueryThreshold time.Duration
}

func NewQueryHook(opts ...Option) *QueryHook {
	h := new(QueryHook)
	for _, opt := range opts {
		opt(h)
	}

	if h.logger == nil {
		h.logger = logger.NewZerologLogger(false)
	}

	if h.slowQueryThreshold == 0 {
		h.slowQueryThreshold = time.Duration(100) * time.Millisecond
	}

	return h
}

func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)
	if duration <= h.slowQueryThreshold && event.Err == nil && !h.debug {
		return
	}

	subLogger := getCtxSubLogger(ctx, h.logger)

	var logEvent logger.LogEvent

	if event.Err != nil {
		if errors.Is(event.Err, sql.ErrNoRows) || errors.Is(event.Err, sql.ErrTxDone) {
			logEvent = subLogger.Info().Err(event.Err)
		} else {
			logEvent = subLogger.Error().Err(event.Err)
		}
	} else if duration > h.slowQueryThreshold {
		logEvent = subLogger.Warn()
	} else {
		logEvent = subLogger.Debug()
	}

	logEvent.
		Field("component", "mysql_db").
		Field("duration_ms", duration.Milliseconds()).
		Field("query", strings.TrimSpace(event.Query)).
		Msgf("SQL %s", event.Operation())
}

func getCtxSubLogger(ctx context.Context, defaultLogger logger.Logger) logger.Logger {
	if subLogger, ok := ctx.Value(constant.SubLoggerCtxKey).(logger.Logger); ok {
		return subLogger
	}

	return defaultLogger
}
