package redisrepository

import (
	"context"
	"goapptemp/internal/shared/exception"
	"net"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

func handleRedisError(err error, operationDesc string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, redis.Nil) {
		return errors.Wrap(exception.ErrNotFound, operationDesc)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return errors.Wrap(exception.ErrTimeout, operationDesc)
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return errors.Wrap(exception.ErrTimeout, operationDesc)
		}

		return errors.Wrap(exception.ErrConnection, operationDesc)
	}

	if errors.Is(err, redis.TxFailedErr) {
		return errors.Wrap(exception.ErrTxFailed, operationDesc)
	}

	return errors.Wrap(err, operationDesc+": redis error occurred")
}
