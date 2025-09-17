package handler

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"

	"github.com/cockroachdb/errors"
	validator "github.com/go-playground/validator/v10"
	echo "github.com/labstack/echo/v4"
)

type AuthHandler struct {
	properties
}

func NewAuthHandler(properties properties) *AuthHandler {
	return &AuthHandler{
		properties: properties,
	}
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	shared.Sanitize(req, nil)

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	user, err := h.service.Auth().Login(ctx, &service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	data := serializer.SerializeUser(user)

	return response.Success(c, "Login success", data)
}
