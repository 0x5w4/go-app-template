package handler

import (
	"fmt"
	"strconv"

	"goapptemp/constant"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared/exception"

	"github.com/labstack/echo/v4"
)

func parseUintParam(c echo.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		msg := fmt.Sprintf("%s is required in URL path", paramName)
		err := exception.New(exception.TypeBadRequest, exception.CodeValidationFailed, msg)

		return 0, exception.WithFieldError(err, paramName, msg)
	}

	id, parseErr := strconv.ParseUint(idStr, 10, 32)
	if parseErr != nil || id == 0 {
		msg := fmt.Sprintf("%s must be a positive integer in URL path", paramName)
		err := exception.Wrap(parseErr, exception.TypeBadRequest, exception.CodeValidationFailed, msg)

		return 0, exception.WithFieldError(err, paramName, msg)
	}

	return uint(id), nil
}

func getAuthArg(c echo.Context) (service.AuthParams, error) {
	arg := c.Get(constant.AuthPayloadCtxKey)
	if arg == nil {
		return service.AuthParams{}, exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderMissing, "no authorization arguments provided")
	}

	authArg, ok := arg.(service.AuthParams)
	if !ok {
		return service.AuthParams{}, exception.New(exception.TypePermissionDenied, exception.CodeAuthHeaderInvalid, "Invalid authorization arguments")
	}

	return authArg, nil
}
