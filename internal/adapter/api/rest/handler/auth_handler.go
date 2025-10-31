package handler

import (
	"goapptemp/constant"
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/internal/shared/token"

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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(RefreshRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "failed to bind data")
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		} else {
			return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "request validation failed")
		}
	}

	token, err := h.service.Auth().Refresh(ctx, &service.RefreshRequest{RefreshToken: req.RefreshToken})
	if err != nil {
		return err
	}

	data := serializer.SerializeToken(token)

	return response.Success(c, "Refresh success", data)
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(LogoutRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "failed to bind data")
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "request validation failed")
	}

	claims, ok := ctx.Value(constant.CtxKeyAuthPayload).(*token.AccessTokenClaims)
	if !ok || claims == nil {
		return exception.New(exception.TypeUnauthorized, exception.CodeUnauthorized, "invalid authorization claims")
	}

	err := h.service.Auth().Logout(ctx, &service.LogoutRequest{
		AccessTokenClaims: claims,
		RefreshToken:      req.RefreshToken,
	})
	if err != nil {
		return err
	}

	return response.Success(c, "Logout success", nil)
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *AuthHandler) ForgetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(ForgetPasswordRequest)

	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "failed to bind data")
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "request validation failed")
	}

	err := h.service.Auth().ForgetPassword(ctx, &service.ForgetPasswordRequest{
		Email: req.Email,
	})
	if err != nil {
		return err
	}

	return response.Success(c, "If your email is registered, you will receive a password reset link.", nil)
}

type VerifyResetTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

func (h *AuthHandler) VerifyResetToken(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(VerifyResetTokenRequest)

	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "failed to bind data")
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "request validation failed")
	}

	err := h.service.Auth().VerifyResetToken(ctx, &service.VerifyResetTokenRequest{
		Token: req.Token,
	})
	if err != nil {
		return err
	}

	return response.Success(c, "Token is valid.", nil)
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(ResetPasswordRequest)

	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "failed to bind data")
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "request validation failed")
	}

	err := h.service.Auth().ResetPassword(ctx, &service.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return err
	}

	return response.Success(c, "Password has been reset successfully.", nil)
}
