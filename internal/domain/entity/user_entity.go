package entity

import (
	"goapptemp/internal/adapter/util"
)

type User struct {
	Base
	RoleIDs   []uint
	Roles     []*Role
	CompanyID uint
	Fullname  string
	Username  string
	Email     string
	Password  string
	Token     *string
}

func (e *User) SetPassword(password string) error {
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return err
	}

	e.Password = hashedPassword

	return nil
}

type UserRole struct {
	UserID uint
	User   *User
	RoleID uint
	Role   *Role
}
