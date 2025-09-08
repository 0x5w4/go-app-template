package entity

type Role struct {
	Base
	PermissionIDs []uint
	Permissions   []*Permission
	Code          string
	Name          string
	Description   *string
	SuperAdmin    bool
}

type RolePermission struct {
	RoleId       uint
	Role         *Role
	PermissionId uint
	Permission   *Permission
}
