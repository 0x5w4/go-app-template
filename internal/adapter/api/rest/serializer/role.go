package serializer

import (
	"goapptemp/internal/domain/entity"
	"time"
)

type RoleResponseData struct {
	ID            uint                      `json:"id"`
	PermissionIDs []uint                    `json:"permission_ids"`
	Permissions   []*PermissionResponseData `json:"permissions"`
	Code          string                    `json:"code"`
	Name          string                    `json:"name"`
	Description   *string                   `json:"description,omitempty"`
	SuperAdmin    bool                      `json:"super_admin"`
	CreatedAt     string                    `json:"created_at,omitempty"`
	UpdatedAt     string                    `json:"updated_at,omitempty"`
}

func SerializeRole(arg *entity.Role) *RoleResponseData {
	if arg == nil {
		return nil
	}

	return &RoleResponseData{
		ID:            arg.ID,
		PermissionIDs: arg.PermissionIDs,
		Permissions:   SerializePermissions(arg.Permissions),
		Code:          arg.Code,
		Name:          arg.Name,
		Description:   arg.Description,
		SuperAdmin:    arg.SuperAdmin,
		CreatedAt:     arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     arg.UpdatedAt.Format(time.RFC3339),
	}
}

func SerializeRoles(arg []*entity.Role) []*RoleResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*RoleResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeRole(arg[i]))
	}

	return res
}
