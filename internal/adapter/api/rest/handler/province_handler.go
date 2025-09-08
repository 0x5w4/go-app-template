package handler

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/adapter/util"
	"goapptemp/internal/adapter/util/exception"
	"goapptemp/internal/domain/service/province"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type FilterProvinceRequest struct {
	IDs     []uint   `query:"ids" validate:"omitempty,dive,gt=0"`
	Names   []string `query:"names" validate:"omitempty,dive,min=2,max=100"`
	Search  string   `query:"search" validate:"omitempty,min=1"`
	Page    int      `query:"page" validate:"omitempty,min=1"`
	PerPage int      `query:"per_page" validate:"omitempty,min=1,max=100"`
}

func (h *Handler) FindProvinces(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(FilterProvinceRequest)
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

	provinces, totalCount, err := h.service.Province().Find(ctx, &province.FindProvincesRequest{
		Filter: &mysql.FilterProvincePayload{
			IDs:     req.IDs,
			Names:   req.Names,
			Search:  req.Search,
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
	if err != nil {
		return err
	}

	list := serializer.SerializeProvinces(provinces)

	pagination := response.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: totalCount,
		TotalPage:  0,
	}
	if req.PerPage > 0 {
		pagination.TotalPage = (totalCount + req.PerPage - 1) / req.PerPage
	}

	return response.Paginate(c, "Find provinces success", list, pagination)
}

func (h *Handler) FindOneProvince(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	province, err := h.service.Province().FindOne(ctx, &province.FindOneProvinceRequest{ProvinceID: id})
	if err != nil {
		return err
	}

	data := serializer.SerializeProvince(province)

	return response.Success(c, "Find one province success", data)
}
