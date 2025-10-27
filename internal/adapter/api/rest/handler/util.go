package handler

import (
	"goapptemp/constant"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared/exception"
	"strconv"

	echo "github.com/labstack/echo/v4"
)

func parseUintParam(c echo.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		msg := paramName + " is required in URL path"
		err := exception.New(exception.TypeBadRequest, exception.CodeValidationFailed, msg)

		return 0, exception.WithFieldError(err, paramName, msg)
	}

	id, parseErr := strconv.ParseUint(idStr, 10, 32)
	if parseErr != nil || id == 0 {
		msg := paramName + " must be a positive integer in URL path"
		err := exception.Wrap(parseErr, exception.TypeBadRequest, exception.CodeValidationFailed, msg)

		return 0, exception.WithFieldError(err, paramName, msg)
	}

	return uint(id), nil
}

func getAuthArg(c echo.Context) (service.AuthParams, error) {
	arg := c.Get(constant.CtxKeyAuthPayload)
	if arg == nil {
		return service.AuthParams{}, exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderMissing, "no authorization arguments provided")
	}

	authArg, ok := arg.(service.AuthParams)
	if !ok {
		return service.AuthParams{}, exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderInvalid, "Invalid authorization arguments")
	}

	return authArg, nil
}
