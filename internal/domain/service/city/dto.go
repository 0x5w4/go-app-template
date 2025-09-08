package city

import "goapptemp/internal/adapter/repository/mysql"

type FindCitiesRequest struct {
	Filter *mysql.FilterCityPayload
}

type FindOneCityRequest struct {
	CityID uint
}
