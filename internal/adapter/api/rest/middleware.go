package rest

import (
	"context"
	"goapptemp/constant"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	apmecho "go.elastic.co/apm/module/apmechov4/v2"
)

func (s *echoServer) setupMiddlewares() {
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	s.echo.Use(s.requestLoggerMiddleware())
	s.echo.Use(apmecho.Middleware())
	s.echo.HTTPErrorHandler = s.httpErrorHandler
}

func (s *echoServer) requestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startTime := time.Now()

			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			c.Set(constant.CtxKeyRequestID, reqID)

			reqLogger := s.logger.NewInstance().Field("request_id", reqID).Logger()
			c.Set(constant.CtxKeySubLogger, reqLogger)

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

func (s *echoServer) authMiddleware(autoDenied bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				if autoDenied {
					return exception.ErrAuthHeaderMissing
				}

				return next(c)
			}

			parts := strings.Fields(authHeader)
			if len(parts) != 2 {
				return errors.Wrapf(exception.ErrAuthHeaderInvalid, "expected 2 parts, got %d", len(parts))
			}

			tokenType, accessToken := parts[0], parts[1]

			if !strings.EqualFold(tokenType, constant.TokenType) {
				return errors.Wrapf(exception.ErrAuthUnsupported, "scheme %q is not %s", tokenType, constant.TokenType)
			}

			claims, verifyErr := s.token.VerifyAccessToken(accessToken)
			if verifyErr != nil {
				return exception.ErrAuthTokenInvalid
			}

			ctx := c.Request().Context()

			isBlacklisted, err := s.redis.CheckTokenBlacklisted(ctx, claims.ID)
			if err != nil {
				return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to verify token")
			}

			if isBlacklisted {
				return exception.ErrAuthTokenBlacklisted
			}

			authParam := service.AuthParams{
				AccessToken:       accessToken,
				AccessTokenClaims: claims,
			}
			c.Set(constant.CtxKeyAuthPayload, authParam)

			return next(c)
		}
	}
}

func (s *echoServer) rateLimitMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			ip := c.RealIP()

			ttl, err := s.redis.GetBlockIPTTL(ctx, ip)
			if err == nil && ttl > 0 {
				retryAfterSeconds := int(ttl.Seconds())
				c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))

				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"message": "Too Many Requests",
				})
			}

			newCtx := context.WithValue(ctx, constant.CtxKeyRequestIP, ip)
			c.SetRequest(c.Request().WithContext(newCtx))

			return next(c)
		}
	}
}
