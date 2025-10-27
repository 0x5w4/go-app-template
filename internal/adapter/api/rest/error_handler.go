package rest

import (
	"goapptemp/constant"
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"net/http"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	apm "go.elastic.co/apm/v2"
)

func (s *echoServer) httpErrorHandler(err error, c echo.Context) {
	requestID, ok := c.Get(constant.CtxKeyRequestID).(string)
	if !ok || requestID == "" {
		s.logger.Warn().Msg("Request ID not found in context, using empty string")
	}

	if apmErr := apm.CaptureError(c.Request().Context(), err); apmErr != nil {
		apmErr.Handled = true
		apmErr.Context.SetHTTPRequest(c.Request())
		apmErr.Send()
	}

	log, ok := c.Get(constant.CtxKeySubLogger).(logger.Logger)
	if !ok || log == nil {
		log = s.logger.NewInstance().Field("request_id", requestID).Logger()
	}

	if err == nil || c.Response().Committed {
		if err != nil {
			log.Error().Msgf("Error handler called after response committed: %+v", err)
		}

		return
	}

	var (
		statusCode  int
		responseMsg string
		errorDetail any
		logMsg      string
		httpErr     *echo.HTTPError
		exMarker    *exception.Exception
	)

	switch {
	case errors.As(err, &httpErr):
		logMsg = "HTTP framework error occurred"

		if errors.As(httpErr.Internal, &exMarker) {
			statusCode, responseMsg, errorDetail = buildErrorPayload(exMarker, httpErr.Code, true, requestID)
		} else {
			statusCode, responseMsg, errorDetail = buildErrorPayload(nil, httpErr.Code, true, requestID)

			var customMsg string
			if mStr, ok := httpErr.Message.(string); ok {
				customMsg = mStr
			} else if mErr, ok := httpErr.Message.(error); ok {
				customMsg = mErr.Error()
			}

			if customMsg != "" && customMsg != http.StatusText(statusCode) {
				responseMsg = customMsg
			}
		}

	case errors.As(err, &exMarker):
		logMsg = "Application error occurred"
		statusCode, responseMsg, errorDetail = buildErrorPayload(exMarker, 0, false, requestID)

	default:
		logMsg = "Unhandled internal error occurred"
		statusCode, responseMsg, errorDetail = buildErrorPayload(nil, http.StatusInternalServerError, true, requestID)
	}

	if statusCode >= http.StatusInternalServerError {
		log.Error().Msgf("%s: Status=%d ResponseMsg='%s' Error: %+v", logMsg, statusCode, responseMsg, err)
	} else {
		log.Warn().Msgf("%s: Status=%d ResponseMsg='%s' Details: %v", logMsg, statusCode, responseMsg, err)
	}

	if err := response.Error(c, statusCode, responseMsg, errorDetail); err != nil {
		log.Error().Err(err).Msg("Failed to send error response")
	}
}

func buildErrorPayload(ex *exception.Exception, initialStatusCode int, forceGeneric bool, requestID string) (statusCode int, message string, errorDetail any) {
	defaultMessage := "An internal server error occurred."
	defaultDetail := map[string]any{"type": string(exception.TypeInternalError), "request_id": requestID}

	if ex != nil {
		switch ex.Type {
		case exception.TypeBadRequest:
			statusCode = http.StatusBadRequest
		case exception.TypeValidationError:
			statusCode = http.StatusUnprocessableEntity
		case exception.TypeUnauthorized, exception.TypeTokenExpired, exception.TypeTokenInvalid, exception.TypeAuthenticationError:
			statusCode = http.StatusUnauthorized
		case exception.TypePermissionDenied, exception.TypeForbidden:
			statusCode = http.StatusForbidden
		case exception.TypeNotFound:
			statusCode = http.StatusNotFound
		case exception.TypeConflict:
			statusCode = http.StatusConflict
		case exception.TypeUnsupportedMediaType:
			statusCode = http.StatusUnsupportedMediaType
		case exception.TypeRateLimitExceeded:
			statusCode = http.StatusTooManyRequests
		case exception.TypeMethodNotAllowed:
			statusCode = http.StatusMethodNotAllowed
		case exception.TypeTimeout, exception.TypeServiceUnavailable, exception.TypeConnectionError, exception.TypeResourceError:
			statusCode = http.StatusServiceUnavailable
		case exception.TypeQueryError, exception.TypeInternalError:
			statusCode = http.StatusInternalServerError
		default:
			if initialStatusCode > 0 {
				statusCode = initialStatusCode
			} else {
				statusCode = http.StatusInternalServerError
			}
		}
	}

	isInternal := statusCode >= http.StatusInternalServerError

	switch {
	case isInternal && (forceGeneric || ex == nil):
		message = defaultMessage
		errorDetail = defaultDetail
	case ex != nil:
		message = ex.Message
		detail := map[string]any{"type": string(ex.Type), "request_id": requestID}

		if ex.Code != "" {
			detail["code"] = ex.Code
		}

		if len(ex.Errors) > 0 {
			detail["details"] = ex.Errors
		}

		errorDetail = detail

	default:
		var defaultType exception.ErrorType

		switch statusCode {
		case http.StatusNotFound:
			message, defaultType = "The requested resource was not found.", exception.TypeNotFound
		case http.StatusMethodNotAllowed:
			message, defaultType = "Method not allowed for this resource.", exception.TypeMethodNotAllowed
		case http.StatusServiceUnavailable:
			message, defaultType = "Service temporarily unavailable.", exception.TypeServiceUnavailable
		default:
			if statusCode >= http.StatusInternalServerError {
				message, defaultType = "An internal server error occurred.", exception.TypeInternalError
			} else {
				message, defaultType = "", ""
			}
		}

		detail := map[string]any{"request_id": requestID}

		if defaultType != "" {
			detail["type"] = string(defaultType)
		} else {
			detail["type"] = "Http Client Error"
		}

		errorDetail = detail
	}

	if message == "" {
		message = http.StatusText(statusCode)
	}

	return statusCode, message, errorDetail
}
