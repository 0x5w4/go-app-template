package province

import "goapptemp/internal/adapter/repository/mysql"

type FindProvincesRequest struct {
	Filter *mysql.FilterProvincePayload
}

type FindOneProvinceRequest struct {
	ProvinceID uint
}
