package serializer

import (
	"goapptemp/internal/domain/entity"
	"time"
)

type CompanyResponseData struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

func SerializeCompany(arg *entity.Company) *CompanyResponseData {
	if arg == nil {
		return nil
	}

	return &CompanyResponseData{
		ID:        arg.ID,
		Name:      arg.Name,
		Icon:      arg.Icon,
		CreatedAt: arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt: arg.UpdatedAt.Format(time.RFC3339),
	}
}

func SerializeCompanies(arg []*entity.Company) []*CompanyResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*CompanyResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeCompany(arg[i]))
	}

	return res
}
