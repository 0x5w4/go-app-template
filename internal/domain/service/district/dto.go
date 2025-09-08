package district

import "goapptemp/internal/adapter/repository/mysql"

type FindDistrictsRequest struct {
	Filter *mysql.FilterDistrictPayload
}

type FindOneDistrictRequest struct {
	DistrictID uint
}
