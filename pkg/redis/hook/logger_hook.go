package redishook

import (
	"context"
	"errors"
	"goapptemp/constant"
	"goapptemp/pkg/logger"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = (*LoggerHook)(nil)

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

func WithSlowThreshold(threshold time.Duration) LoggerOption {
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

func (h *LoggerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *LoggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		startTime := time.Now()

		err := next(ctx, cmd)

		duration := time.Since(startTime)

		if duration <= h.slowQueryThreshold && err == nil && !h.debug {
			return nil
		}

		var subLogger logger.Logger
		if l, ok := ctx.Value(constant.CtxKeySubLogger).(logger.Logger); ok {
			subLogger = l
		} else {
			subLogger = h.logger
		}

		var logEvent logger.LogEvent

		switch {
		case err != nil:
			if errors.Is(err, redis.Nil) {
				logEvent = subLogger.Info().Err(err)
			} else {
				logEvent = subLogger.Error().Err(err)
			}
		case duration > h.slowQueryThreshold:
			logEvent = subLogger.Warn()
		default:
			logEvent = subLogger.Debug()
		}

		logEvent.
			Field("component", "redis_db").
			Field("duration_ms", duration.Milliseconds()).
			Field("command", cmd.Name()).
			Field("args", commandArgsToString(cmd.Args())).
			Msgf("Redis %s", cmd.Name())

		return err
	}
}

func (h *LoggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		startTime := time.Now()

		err := next(ctx, cmds)

		duration := time.Since(startTime)

		var (
			pipelineFailed bool
			firstErr       error
		)

		if err == nil {
			for _, cmd := range cmds {
				if cmd.Err() != nil && !errors.Is(cmd.Err(), redis.Nil) {
					pipelineFailed = true
					firstErr = cmd.Err()

					break
				}
			}
		} else {
			pipelineFailed = true
			firstErr = err
		}

		if duration <= h.slowQueryThreshold && !pipelineFailed && !h.debug {
			return nil
		}

		var subLogger logger.Logger
		if l, ok := ctx.Value(constant.CtxKeySubLogger).(logger.Logger); ok {
			subLogger = l
		} else {
			subLogger = h.logger
		}

		var logEvent logger.LogEvent

		switch {
		case pipelineFailed:
			logEvent = subLogger.Error().Err(firstErr)
		case duration > h.slowQueryThreshold:
			logEvent = subLogger.Warn()
		default:
			logEvent = subLogger.Debug()
		}

		logEvent.
			Field("component", "redis_db").
			Field("duration_ms", duration.Milliseconds()).
			Field("num_commands", len(cmds)).
			Msgf("Redis Pipeline (%d commands)", len(cmds))

		return err
	}
}
