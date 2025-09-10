package serializer

import (
	"goapptemp/internal/domain/entity"
	"time"
)

type UserResponseData struct {
	ID        uint                `json:"id"`
	RoleIDs   []uint              `json:"role_ids"`
	Roles     []*RoleResponseData `json:"roles"`
	Email     string              `json:"email"`
	Username  string              `json:"username"`
	Fullname  string              `json:"fullname"`
	Token     *TokenResponseData  `json:"token,omitempty"`
	CreatedAt string              `json:"created_at,omitempty"`
	UpdatedAt string              `json:"updated_at,omitempty"`
}

func SerializeUser(arg *entity.User) *UserResponseData {
	if arg == nil {
		return nil
	}

	return &UserResponseData{
		ID:        arg.ID,
		RoleIDs:   arg.RoleIDs,
		Roles:     SerializeRoles(arg.Roles),
		Email:     arg.Email,
		Username:  arg.Username,
		Fullname:  arg.Fullname,
		CreatedAt: arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt: arg.UpdatedAt.Format(time.RFC3339),
		Token:     SerializeToken(arg.Token),
	}
}

func SerializeUsers(arg []*entity.User) []*UserResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*UserResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeUser(arg[i]))
	}

	return res
}
