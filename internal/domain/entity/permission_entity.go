package entity

type Permission struct {
	Base
	Code        string
	Name        string
	Description *string
}
