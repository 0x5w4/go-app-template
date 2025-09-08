package server

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/util/constant"
	"goapptemp/internal/adapter/util/exception"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"go.elastic.co/apm/v2"
)

var (
	ErrAuthHeaderMissing = exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderMissing, "Authorization header not provided")
	ErrAuthHeaderInvalid = exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderInvalid, "Invalid authorization header format")
	ErrAuthUnsupported   = exception.New(exception.TypePermissionDenied, exception.CodeAuthUnsupported, "Unsupported authorization type")
	ErrAuthTokenInvalid  = exception.New(exception.TypeBadRequest, exception.CodeTokenInvalid, "Invalid or expired token")
)

func (s *server) httpErrorHandler(err error, c echo.Context) {
	requestID, ok := c.Get(constant.RequestIDCtxKey).(string)
	if !ok || requestID == "" {
		s.logger.Warn().Msg("Request ID not found in context, using empty string")
	}

	if apmErr := apm.CaptureError(c.Request().Context(), err); apmErr != nil {
		apmErr.Handled = true
		apmErr.Context.SetHTTPRequest(c.Request())
		apmErr.Send()
	}

	log, ok := c.Get(constant.SubLoggerCtxKey).(logger.Logger)
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

func mapExceptionTypeToStatusCode(exType exception.ErrorType) int {
	switch exType {
	case exception.TypeBadRequest:
		return http.StatusBadRequest
	case exception.TypeValidationError:
		return http.StatusUnprocessableEntity
	case exception.TypeUnauthorized, exception.TypeTokenExpired, exception.TypeTokenInvalid, exception.TypeAuthenticationError:
		return http.StatusUnauthorized
	case exception.TypePermissionDenied, exception.TypeForbidden:
		return http.StatusForbidden
	case exception.TypeNotFound:
		return http.StatusNotFound
	case exception.TypeConflict:
		return http.StatusConflict
	case exception.TypeUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	case exception.TypeRateLimitExceeded:
		return http.StatusTooManyRequests
	case exception.TypeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case exception.TypeTimeout, exception.TypeServiceUnavailable, exception.TypeConnectionError, exception.TypeResourceError:
		return http.StatusServiceUnavailable
	case exception.TypeQueryError, exception.TypeInternalError:
		return http.StatusInternalServerError
	default:
		return 0
	}
}
func getDefaultPayloadForStatus(statusCode int) (message string, errType exception.ErrorType) {
	switch statusCode {
	case http.StatusNotFound:
		return "The requested resource was not found.", exception.TypeNotFound
	case http.StatusMethodNotAllowed:
		return "Method not allowed for this resource.", exception.TypeMethodNotAllowed
	case http.StatusServiceUnavailable:
		return "Service temporarily unavailable.", exception.TypeServiceUnavailable
	default:
		if statusCode >= http.StatusInternalServerError {
			return "An internal server error occurred.", exception.TypeInternalError
		}

		return "", ""
	}
}

func buildErrorPayload(ex *exception.Exception, initialStatusCode int, forceGeneric bool, requestID string) (statusCode int, message string, errorDetail any) {
	defaultMessage := "An internal server error occurred."
	defaultDetail := map[string]any{"type": string(exception.TypeInternalError), "request_id": requestID}

	if ex != nil {
		statusCode = mapExceptionTypeToStatusCode(ex.Type)
	}

	if statusCode == 0 {
		if initialStatusCode > 0 {
			statusCode = initialStatusCode
		} else {
			statusCode = http.StatusInternalServerError
		}
	}

	isInternal := statusCode >= http.StatusInternalServerError

	switch {
	case isInternal && (forceGeneric || ex == nil):
		message = defaultMessage
		errorDetail = defaultDetail
	case ex != nil:
		message = ex.Message
		errorDetail = buildExceptionDetail(ex, requestID)
	default:
		var defaultType exception.ErrorType
		message, defaultType = getDefaultPayloadForStatus(statusCode)
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

func buildExceptionDetail(ex *exception.Exception, requestID string) map[string]any {
	errMap := map[string]any{"type": string(ex.Type), "request_id": requestID}
	if ex.Code != "" {
		errMap["code"] = ex.Code
	}

	if len(ex.Errors) > 0 {
		errMap["details"] = ex.Errors
	}

	return errMap
}
