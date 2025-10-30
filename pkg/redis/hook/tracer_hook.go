package redishook

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"
	apm "go.elastic.co/apm/v2"
)

var _ redis.Hook = (*TracerHook)(nil)

type TracerHook struct{}

func NewTracerHook() *TracerHook {
	return &TracerHook{}
}

func (h *TracerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *TracerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		span, ctx := apm.StartSpan(ctx, "Redis "+cmd.Name(), "db.query")

		ctx = apm.ContextWithSpan(ctx, span)
		defer span.End()

		err := next(ctx, cmd)

		args := commandArgsToString(cmd.Args())
		if len(args) > 100 {
			args = args[:100] + "..."
		}

		span.Context.SetLabel("redis.command", cmd.Name())
		span.Context.SetLabel("redis.args", args)
		span.Context.SetLabel("db.system", "redis")

		if err != nil && !errors.Is(err, redis.Nil) {
			apm.CaptureError(ctx, err).Send()

			span.Outcome = "failure"
		} else {
			span.Outcome = "success"
		}

		return err
	}
}

func (h *TracerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		span, ctx := apm.StartSpan(ctx, "Redis Pipeline", "db.query")

		ctx = apm.ContextWithSpan(ctx, span)
		defer span.End()

		err := next(ctx, cmds)

		var cmdNames []string
		for _, cmd := range cmds {
			cmdNames = append(cmdNames, cmd.Name())
		}

		span.Context.SetLabel("redis.commands", strings.Join(cmdNames, ", "))
		span.Context.SetLabel("redis.num_commands", len(cmds))
		span.Context.SetLabel("db.system", "redis")

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

		if pipelineFailed {
			apm.CaptureError(ctx, firstErr).Send()

			span.Outcome = "failure"
		} else {
			span.Outcome = "success"
		}

		return err
	}
}
