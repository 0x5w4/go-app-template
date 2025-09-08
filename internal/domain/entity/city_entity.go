package entity

type City struct {
	ID         uint
	ProvinceID uint
	Province   *Province
	Name       string
}
