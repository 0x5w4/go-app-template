package entity

type District struct {
	ID     uint
	CityID uint
	City   *City
	Name   string
}
