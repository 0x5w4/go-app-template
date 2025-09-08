package serializer

import (
	"goapptemp/internal/domain/entity"
	"time"
)

type PermissionResponseData struct {
	ID          uint    `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

func SerializePermission(arg *entity.Permission) *PermissionResponseData {
	if arg == nil {
		return nil
	}

	return &PermissionResponseData{
		ID:          arg.ID,
		Code:        arg.Code,
		Name:        arg.Name,
		Description: arg.Description,
		CreatedAt:   arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   arg.UpdatedAt.Format(time.RFC3339),
	}
}

func SerializePermissions(arg []*entity.Permission) []*PermissionResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*PermissionResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializePermission(arg[i]))
	}

	return res
}
