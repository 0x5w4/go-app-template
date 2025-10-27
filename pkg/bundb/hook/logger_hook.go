package hook

import (
	"context"
	"database/sql"
	"goapptemp/constant"
	"goapptemp/pkg/logger"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/uptrace/bun"
)

var _ bun.QueryHook = (*LoggerHook)(nil)

type LoggerOption func(h *LoggerHook)

func WithLogger(logger logger.Logger) LoggerOption {
	return func(h *LoggerHook) {
		h.logger = logger
	}
}

func WithDebug(debug bool) LoggerOption {
	return func(h *LoggerHook) {
		h.debug = debug
	}
}

func WithSlowQueryThreshold(threshold time.Duration) LoggerOption {
	return func(h *LoggerHook) {
		h.slowQueryThreshold = threshold
	}
}

type LoggerHook struct {
	logger             logger.Logger
	debug              bool
	slowQueryThreshold time.Duration
}

func NewLoggerHook(opts ...LoggerOption) *LoggerHook {
	h := new(LoggerHook)
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

func (h *LoggerHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *LoggerHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)
	if duration <= h.slowQueryThreshold && event.Err == nil && !h.debug {
		return
	}

	var subLogger logger.Logger
	if l, ok := ctx.Value(constant.CtxKeySubLogger).(logger.Logger); ok {
		subLogger = l
	} else {
		subLogger = h.logger
	}

	var logEvent logger.LogEvent

	switch {
	case event.Err != nil:
		if errors.Is(event.Err, sql.ErrNoRows) || errors.Is(event.Err, sql.ErrTxDone) {
			logEvent = subLogger.Info().Err(event.Err)
		} else {
			logEvent = subLogger.Error().Err(event.Err)
		}
	case duration > h.slowQueryThreshold:
		logEvent = subLogger.Warn()
	default:
		logEvent = subLogger.Debug()
	}

	logEvent.
		Field("component", "mysql_db").
		Field("duration_ms", duration.Milliseconds()).
		Field("query", strings.TrimSpace(event.Query)).
		Msgf("SQL %s", event.Operation())
}
