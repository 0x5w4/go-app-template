package handler

import (
	"goapptemp/internal/adapter/api/rest/response"
	"goapptemp/internal/adapter/api/rest/serializer"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"

	"github.com/cockroachdb/errors"
	validator "github.com/go-playground/validator/v10"
	echo "github.com/labstack/echo/v4"
)

type RoleHandler struct {
	properties
}

func NewRoleHandler(properties properties) *RoleHandler {
	return &RoleHandler{
		properties: properties,
	}
}

type CreateRole struct {
	PermissionIDs []uint  `json:"permission_ids"        validate:"required,dive,required,gt=0"`
	Code          string  `json:"code"                  validate:"required,min=2,max=50,code_chars_allowed"`
	Name          string  `json:"name"                  validate:"required,min=2,max=100"`
	Description   *string `json:"description,omitempty" validate:"max=255"`
	SuperAdmin    bool    `json:"super_admin"`
}

type CreateRoleRequest struct {
	Role CreateRole `json:"role" validate:"required"`
}

func (h *RoleHandler) CreateRole(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(CreateRoleRequest)
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

	role, err := h.service.Role().Create(ctx,
		&service.CreateRoleRequest{
			AuthParams: &authArg,
			Role: &entity.Role{
				PermissionIDs: req.Role.PermissionIDs,
				Name:          req.Role.Name,
				Code:          req.Role.Code,
				Description:   req.Role.Description,
				SuperAdmin:    req.Role.SuperAdmin,
			},
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeRole(role)

	return response.Success(c, "Create role success", data)
}

type FilterRoleRequest struct {
	IDs        []uint   `validate:"omitempty,dive,gt=0"                            query:"ids"`
	Codes      []string `validate:"omitempty,dive,min=2,max=50,code_chars_allowed" query:"codes"`
	Names      []string `validate:"omitempty,dive,min=2,max=100"                   query:"names"`
	SuperAdmin *bool    `validate:"omitempty"                                      query:"super_admin"`
	Search     string   `validate:"omitempty,min=1"                                query:"search"`
	Page       int      `validate:"omitempty,min=1"                                query:"page"`
	PerPage    int      `validate:"omitempty,min=1,max=100"                        query:"per_page"`
}

func (h *RoleHandler) FindRoles(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(FilterRoleRequest)
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

	roles, totalCount, err := h.service.Role().Find(ctx,
		&service.FindRolesRequest{
			AuthParams: &authArg,
			Filter: &mysqlrepository.FilterRolePayload{
				IDs:        req.IDs,
				Names:      req.Names,
				Codes:      req.Codes,
				SuperAdmin: req.SuperAdmin,
				Search:     req.Search,
				Page:       req.Page,
				PerPage:    req.PerPage,
			},
		})
	if err != nil {
		return err
	}

	list := serializer.SerializeRoles(roles)

	pagination := response.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: totalCount,
		TotalPage:  0,
	}
	if req.PerPage > 0 {
		pagination.TotalPage = (totalCount + req.PerPage - 1) / req.PerPage
	}

	return response.Paginate(c, "Find roles success", list, pagination)
}

func (h *RoleHandler) FindOneRole(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	role, err := h.service.Role().FindOne(ctx,
		&service.FindOneRoleRequest{
			AuthParams: &authArg,
			RoleID:     id,
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeRole(role)

	return response.Success(c, "Find role success", data)
}

type UpdateRole struct {
	ID            uint    `validate:"required,gt=0"        param:"id"`
	PermissionIDs []*uint `json:"permission_ids,omitempty" validate:"dive,required,gt=0"`
	Code          *string `json:"code,omitempty"           validate:"min=2,max=50,code_chars_allowed"`
	Name          *string `json:"name,omitempty"           validate:"min=2,max=100"`
	Description   *string `json:"description,omitempty"    validate:"max=255"`
	SuperAdmin    *bool   `json:"super_admin,omitempty"`
}

type UpdateRoleRequest struct {
	Role UpdateRole `json:"role" validate:"required"`
}

func (h *RoleHandler) UpdateRole(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(UpdateRoleRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	shared.Sanitize(req, nil)

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	req.Role.ID = id
	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	role, err := h.service.Role().Update(ctx,
		&service.UpdateRoleRequest{
			AuthParams: &authArg,
			Update: &mysqlrepository.UpdateRolePayload{
				ID:            req.Role.ID,
				PermissionIDs: req.Role.PermissionIDs,
				Name:          req.Role.Name,
				Code:          req.Role.Code,
				Description:   req.Role.Description,
				SuperAdmin:    req.Role.SuperAdmin,
			},
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeRole(role)

	return response.Success(c, "Update role success", data)
}

func (h *RoleHandler) DeleteRole(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	err = h.service.Role().Delete(ctx,
		&service.DeleteRoleRequest{
			AuthParams: &authArg,
			RoleID:     id,
		})
	if err != nil {
		return err
	}

	return response.Success(c, "Delete role success", nil)
}
