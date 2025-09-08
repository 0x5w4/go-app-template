package entity

import "time"

type Company struct {
	Base
	Name          string
	Icon          *string
	IconUpdatedAt *time.Time
	AdminID       uint
}
