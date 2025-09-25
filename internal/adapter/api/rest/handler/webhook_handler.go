package handler

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"strconv"

	"github.com/cockroachdb/errors"
	validator "github.com/go-playground/validator/v10"
	echo "github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	properties
}

func NewWebhookHandler(properties properties) *WebhookHandler {
	return &WebhookHandler{
		properties: properties,
	}
}

type UpdateIconRequest struct {
	ID   uint   `validate:"required,gt=0"                        query:"id"`
	Type string `validate:"required,oneof=client group merchant" query:"type"`
	Link string `json:"link"                                     validate:"required,url"`
}

func (h *WebhookHandler) UpdateIcon(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(UpdateIconRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	shared.Sanitize(req, nil)

	idStr := c.QueryParam("id")
	if idStr == "" {
		err := errors.New("ID is required")
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "ID is required")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "ID must be a positive integer")
	}

	req.ID = uint(id)
	req.Type = c.QueryParam("type")

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	h.logger.Debug().
		Field("id", req.ID).
		Field("type", req.Type).
		Field("link_present", req.Link != "").
		Msg("Webhook UpdateIcon request received and validated")

	err = h.service.Webhook().UpdateIcon(ctx, &service.UpdateIconRequest{
		ID:   req.ID,
		Link: req.Link,
		Type: req.Type,
	})
	if err != nil {
		return err
	}

	return response.Success(c, "Update icon success", nil)
}
