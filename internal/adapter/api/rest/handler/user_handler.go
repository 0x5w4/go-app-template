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

type UserHandler struct {
	properties
}

func NewUserHandler(properties properties) *UserHandler {
	return &UserHandler{
		properties: properties,
	}
}

type CreateUser struct {
	RoleIDs  []uint `json:"role_ids" validate:"required,dive,required,gt=0"`
	Fullname string `json:"fullname" validate:"required,min=3,max=100"`
	Username string `json:"username" validate:"required,min=3,max=100,username_chars_allowed"`
	Email    string `json:"email"    validate:"required,email,min=3,max=100"`
	Password string `json:"password" validate:"required,password,max=200"`
}

type CreateUserRequest struct {
	User CreateUser `json:"user" validate:"required"`
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(CreateUserRequest)
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

	user, err := h.service.User().Create(ctx, &service.CreateUserRequest{
		AuthParams: &authArg,
		User: &entity.User{
			RoleIDs:  req.User.RoleIDs,
			Fullname: req.User.Fullname,
			Email:    req.User.Email,
			Username: req.User.Username,
			Password: req.User.Password,
		},
	})
	if err != nil {
		return err
	}

	data := serializer.SerializeUser(user)

	return response.Success(c, "Create user success", data)
}

type FilterUserRequest struct {
	IDs       []uint   `validate:"omitempty,dive,gt=0"                           query:"ids"`
	Usernames []string `validate:"omitemptymin=3,max=100,username_chars_allowed" query:"usernames"`
	Emails    []string `validate:"email,min=3,max=100"                           query:"emails"`
	Search    string   `validate:"omitempty,min=1"                               query:"search"`
	Page      int      `validate:"omitempty,min=1"                               query:"page"`
	PerPage   int      `validate:"omitempty,min=1,max=100"                       query:"per_page"`
}

func (h *UserHandler) FindUsers(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(FilterUserRequest)
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

	users, totalCount, err := h.service.User().Find(ctx,
		&service.FindUserRequest{
			AuthParams: &authArg,
			UserFilter: &mysqlrepository.FilterUserPayload{
				IDs:       req.IDs,
				Usernames: req.Usernames,
				Emails:    req.Emails,
				Search:    req.Search,
				Page:      req.Page,
				PerPage:   req.PerPage,
			},
		})
	if err != nil {
		return err
	}

	list := serializer.SerializeUsers(users)

	pagination := response.Pagination{
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalCount: totalCount,
		TotalPage:  0,
	}
	if req.PerPage > 0 {
		pagination.TotalPage = (totalCount + req.PerPage - 1) / req.PerPage
	}

	return response.Paginate(c, "Find users success", list, pagination)
}

func (h *UserHandler) FindOneUser(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	user, err := h.service.User().FindOne(ctx,
		&service.FindOneUserRequest{
			AuthParams: &authArg,
			UserID:     id,
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeUser(user)

	return response.Success(c, "Find user success", data)
}

type UpdateUser struct {
	ID       uint    `validate:"required,gt=0"  param:"id"`
	RoleIDs  []*uint `json:"role_ids,omitempty" validate:"dive,required,gt=0"`
	Email    *string `json:"email,omitempty"    validate:"email,min=3,max=100"`
	Username *string `json:"username,omitempty" validate:"min=3,max=100,username_chars_allowed"`
	Password *string `json:"password,omitempty" validate:"password,max=200"`
	Fullname *string `json:"fullname,omitempty" validate:"min=3,max=100"`
}

type UpdateUserRequest struct {
	User UpdateUser `json:"user" validate:"required"`
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	req := new(UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Failed to bind data")
	}

	shared.Sanitize(req, nil)

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	req.User.ID = id
	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return exception.FromValidationErrors(req, validationErrors)
		}

		return exception.Wrap(err, exception.TypeBadRequest, exception.CodeValidationFailed, "Request validation failed")
	}

	user, err := h.service.User().Update(ctx,
		&service.UpdateUserRequest{
			AuthParams: &authArg,
			Update: &mysqlrepository.UpdateUserPayload{
				ID:       req.User.ID,
				RoleIDs:  req.User.RoleIDs,
				Fullname: req.User.Fullname,
				Email:    req.User.Email,
				Username: req.User.Username,
				Password: req.User.Password,
			},
		})
	if err != nil {
		return err
	}

	data := serializer.SerializeUser(user)

	return response.Success(c, "Update user success", data)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()

	authArg, err := getAuthArg(c)
	if err != nil {
		return err
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	err = h.service.User().Delete(ctx,
		&service.DeleteUserRequest{
			AuthParams: &authArg,
			UserID:     id,
		})
	if err != nil {
		return err
	}

	return response.Success(c, "Delete user success", nil)
}
