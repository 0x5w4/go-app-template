package server

import (
	"net/http"
	"strings"
	"time"

	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/util/constant"
	"goapptemp/internal/domain/service/auth"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	apmecho "go.elastic.co/apm/module/apmechov4/v2"
)

func (s *server) setupMiddleware() {
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // TODO: Restrict in production
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	s.echo.Use(s.requestLoggerMiddleware())
	s.echo.Use(apmecho.Middleware())
	s.echo.HTTPErrorHandler = s.httpErrorHandler
}

func (s *server) requestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startTime := time.Now()

			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			c.Set(constant.RequestIDCtxKey, reqID)

			reqLogger := s.logger.NewInstance().Field("request_id", reqID).Logger()
			c.Set(constant.SubLoggerCtxKey, reqLogger)

			req := c.Request()
			err := next(c)
			res := c.Response()
			status := res.Status

			var logEvent logger.LogEvent

			switch {
			case status >= 500:
				logEvent = reqLogger.Error()
				if err != nil {
					logEvent = logEvent.Err(err)
				}
			case status >= 400:
				logEvent = reqLogger.Warn()
				if err != nil {
					logEvent = logEvent.Err(err)
				}
			default:
				logEvent = reqLogger.Info()
			}

			logEvent.
				Field("protocol", req.Proto).
				Field("remote_ip", c.RealIP()).
				Field("host", req.Host).
				Field("method", req.Method).
				Field("uri", req.RequestURI).
				Field("request_id", reqID).
				Field("status", status).
				Field("duration_ms", time.Since(startTime).Milliseconds()).
				Field("response_size", res.Size).
				Field("user_agent", req.UserAgent()).
				Msg("HTTP request processed")

			return err
		}
	}
}

func (s *server) authMiddleware(autoDenied bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				if autoDenied {
					return ErrAuthHeaderMissing
				}

				return next(c)
			}

			parts := strings.Fields(authHeader)
			if len(parts) != 2 {
				return errors.Wrapf(ErrAuthHeaderInvalid, "expected 2 parts, got %d", len(parts))
			}

			authType := strings.ToLower(parts[0])
			tokenStr := parts[1]

			if authType != "bearer" {
				return errors.Wrapf(ErrAuthUnsupported, "scheme %q is not bearer", authType)
			}

			claims, verifyErr := s.handler.VerifyToken(tokenStr)
			if verifyErr != nil {
				return verifyErr
			}

			authParam := auth.AuthParams{
				Token:  tokenStr,
				Claims: claims,
			}
			c.Set(constant.AuthPayloadCtxKey, authParam)

			return next(c)
		}
	}
}
