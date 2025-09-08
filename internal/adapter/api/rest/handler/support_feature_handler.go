package handler

import (
	"io"
	"strconv"
	"strings"

	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/adapter/util"
	"goapptemp/internal/adapter/util/exception"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/supportfeature"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CreateSupportFeature struct {
	Name     string `json:"name" validate:"required,min=2,max=32,alpha_space"`
	Key      string `json:"key" validate:"required,min=2,max=32,username_chars_allowed"`
	IsActive bool   `json:"is_active"`
}

type CreateSupportFeatureRequest struct {
	SupportFeature CreateSupportFeature `json:"help_service" validate:"required"`
}

type BulkCreateSupportFeatureRequest struct {
	SupportFeatures []CreateSupportFeature `json:"help_services" validate:"required,dive"`
}

type UpdateSupportFeature struct {
	ID       uint    `param:"id" validate:"required,gt=0"`
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=32,alpha_space"`
	Key      *string `json:"key,omitempty" validate:"omitempty,min=2,max=32,username_chars_allowed"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type UpdateSupportFeatureRequest struct {
	SupportFeature UpdateSupportFeature `json:"help_service" validate:"required"`
}

type FilterSupportFeatureRequest struct {
	IDs      []uint   `query:"ids" validate:"omitempty,dive,gt=0"`
	Codes    []string `query:"codes" validate:"omitempty,dive,min=2,max=50,alphanum"`
	Names    []string `query:"names" validate:"omitempty,dive,min=2,max=32,alpha_space"`
	Keys     []string `query:"keys" validate:"omitempty,dive,min=2,max=32,username_chars_allowed"`
	IsActive *bool    `query:"is_active" validate:"omitempty"`
	Search   string   `query:"search" validate:"omitempty,min=1"`
	Page     int      `query:"page" validate:"omitempty,min=1"`
	PerPage  int      `query:"per_page" validate:"omitempty,min=1,max=100"`
}

func (h *Handler) CreateSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(CreateSupportFeatureRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	util.Sanitize(req)

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	supportFeature, err := h.service.SupportFeature().Create(ctx,
		&supportfeature.CreateSupportFeatureRequest{
			AuthParams: &authArg,
			SupportFeature: &entity.SupportFeature{
				Name:     strings.TrimSpace(req.SupportFeature.Name),
				Key:      req.SupportFeature.Key,
				IsActive: req.SupportFeature.IsActive,
			},
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeSupportFeature(supportFeature)

	return response.Success(c, "Create help service success", data)
}

func (h *Handler) BulkCreateSupportFeatures(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(BulkCreateSupportFeatureRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	util.Sanitize(req)

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	if len(req.SupportFeatures) == 0 {
		return exception.New(exception.TypeBadRequest, exception.CodeValidationFailed, "At least one help service is required")
	}

	if len(req.SupportFeatures) > 300 {
		return exception.New(exception.TypeBadRequest, exception.CodeValidationFailed, "Maximum 300 help services can be created at once")
	}

	sfs := make([]*entity.SupportFeature, 0, len(req.SupportFeatures))

	for i := range req.SupportFeatures {
		sf := &entity.SupportFeature{
			Name:     req.SupportFeatures[i].Name,
			Key:      req.SupportFeatures[i].Key,
			IsActive: req.SupportFeatures[i].IsActive,
		}
		sfs = append(sfs, sf)
	}

	supportFeatures, err := h.service.SupportFeature().BulkCreate(ctx,
		&supportfeature.BulkCreateSupportFeatureRequest{
			AuthParams:      &authArg,
			SupportFeatures: sfs,
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeSupportFeatures(supportFeatures)

	return response.Success(c, "Bulk create help service success", data)
}

func (h *Handler) FindSupportFeatures(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(FilterSupportFeatureRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind parameters")
	}

	util.Sanitize(req)

	if req.Page <= 0 {
		req.Page = 1
	}

	if req.PerPage <= 0 {
		req.PerPage = 10
	} else if req.PerPage > 100 {
		req.PerPage = 100
	}

	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Invalid query parameters")
	}

	supportFeatures, totalCount, err := h.service.SupportFeature().Find(ctx,
		&supportfeature.FindSupportFeaturesRequest{
			AuthParams: &authArg,
			Filter: &mysql.FilterSupportFeaturePayload{
				IDs:      req.IDs,
				Codes:    req.Codes,
				Names:    req.Names,
				Keys:     req.Keys,
				IsActive: req.IsActive,
				Search:   req.Search,
				Page:     req.Page,
				PerPage:  req.PerPage,
			},
		})
	if err != nil {
		return err
	}

	list := serializer.SerializeSupportFeatures(supportFeatures)

	pagination := response.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: totalCount,
		TotalPage:  0,
	}
	if req.PerPage > 0 {
		pagination.TotalPage = (totalCount + req.PerPage - 1) / req.PerPage
	}

	return response.Paginate(c, "Find help services success", list, pagination)
}

func (h *Handler) FindOneSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	supportFeature, err := h.service.SupportFeature().FindOne(ctx,
		&supportfeature.FindOneSupportFeatureRequest{
			AuthParams:       &authArg,
			SupportFeatureID: id,
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeSupportFeature(supportFeature)

	return response.Success(c, "Find one help service success", data)
}

func (h *Handler) UpdateSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(UpdateSupportFeatureRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	util.Sanitize(req)

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	req.SupportFeature.ID = id
	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	supportFeature, err := h.service.SupportFeature().Update(ctx,
		&supportfeature.UpdateSupportFeatureRequest{
			AuthParams: &authArg,
			Update: &mysql.UpdateSupportFeaturePayload{
				ID:       req.SupportFeature.ID,
				Name:     req.SupportFeature.Name,
				Key:      req.SupportFeature.Key,
				IsActive: req.SupportFeature.IsActive,
			},
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeSupportFeature(supportFeature)

	return response.Success(c, "Update help service success", data)
}

func (h *Handler) DeleteSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	err = h.service.SupportFeature().Delete(ctx,
		&supportfeature.DeleteSupportFeatureRequest{
			AuthParams:       &authArg,
			SupportFeatureID: id,
		})
	if err != nil {
		return err
	}

	return response.Success(c, "Delete help service success", nil)
}

func (h *Handler) IsSupportFeatureDeletable(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	isDeletable, err := h.service.SupportFeature().IsDeletable(ctx,
		&supportfeature.IsDeletableSupportFeatureRequest{
			AuthParams:       &authArg,
			SupportFeatureID: id,
		})
	if err != nil {
		return err
	}

	type data struct {
		IsDeletable bool `json:"is_deletable"`
	}

	return response.Success(c, "Check if help service is deletable success", &data{IsDeletable: isDeletable})
}

func (h *Handler) ImportPreviewSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to get file from form")
	}

	if file == nil {
		return exception.New(exception.TypeBadRequest, exception.CodeValidationFailed, "file is required")
	}

	data, err := h.service.SupportFeature().ImportPreview(ctx,
		&supportfeature.ImportPreviewSupportFeatureRequest{
			AuthParams: &authArg,
			File:       file,
		})
	if err != nil {
		return err
	}

	return response.Success(c, "Import preview success", serializer.SerializeSupportFeaturePreviews(data))
}

func (h *Handler) TemplateImportSupportFeature(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	fileData, err := h.service.SupportFeature().TemplateImport(ctx,
		&supportfeature.TemplateImportSupportFeatureRequest{
			AuthParams: &authArg,
		})
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, fileData.MIMEType)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(fileData.Size, 10))
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+strconv.Quote(fileData.Filename))

	if _, err := io.Copy(c.Response().Writer, fileData.Content); err != nil {
		if h.logger != nil {
			h.logger.Error().Err(err).Msg("Failed to write Excel file")
		}

		return err
	}

	return nil
}
