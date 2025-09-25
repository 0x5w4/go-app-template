package response

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func Success(c echo.Context, message string, data any) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c echo.Context, statusCode int, message string, errDetail any) error {
	return c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   errDetail,
	})
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPage  int `json:"total_page"`
	TotalCount int `json:"total_count"`
}

type PaginatedData struct {
	List       any        `json:"list"`
	Pagination Pagination `json:"pagination"`
}

func Paginate(c echo.Context, message string, list any, pagination Pagination) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data: PaginatedData{
			List:       list,
			Pagination: pagination,
		},
	})
}
