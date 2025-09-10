package handler

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type DistrictHandler struct {
	properties
}

func NewDistrictHandler(properties properties) *DistrictHandler {
	return &DistrictHandler{
		properties: properties,
	}
}

type FilterDistrictRequest struct {
	IDs     []uint   `query:"ids" validate:"omitempty,dive,gt=0"`
	CityIDs []uint   `query:"city_ids" validate:"omitempty,dive,gt=0"`
	Names   []string `query:"names" validate:"omitempty,dive,min=2,max=100"`
	Search  string   `query:"search" validate:"omitempty,min=1"`
	Page    int      `query:"page" validate:"omitempty,min=1"`
	PerPage int      `query:"per_page" validate:"omitempty,min=1,max=100"`
}

func (h *DistrictHandler) FindDistricts(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(FilterDistrictRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind parameters")
	}

	shared.Sanitize(req, nil)

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

	districts, totalCount, err := h.service.District().Find(ctx, &service.FindDistrictsRequest{
		Filter: &mysql.FilterDistrictPayload{
			IDs:     req.IDs,
			CityIDs: req.CityIDs,
			Names:   req.Names,
			Search:  req.Search,
			Page:    req.Page,
			PerPage: req.PerPage,
		},
	})
	if err != nil {
		return err
	}

	list := serializer.SerializeDistricts(districts)

	pagination := response.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: totalCount,
		TotalPage:  0,
	}
	if req.PerPage > 0 {
		pagination.TotalPage = (totalCount + req.PerPage - 1) / req.PerPage
	}

	return response.Paginate(c, "Find districts success", list, pagination)
}

func (h *DistrictHandler) FindOneDistrict(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	district, err := h.service.District().FindOne(ctx, &service.FindOneDistrictRequest{DistrictID: id})
	if err != nil {
		return err
	}

	data := serializer.SerializeDistrict(district)

	return response.Success(c, "Find one district success", data)
}
