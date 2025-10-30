package mysqlrepository

import (
	"context"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ DistrictRepository = (*districtRepository)(nil)

type DistrictRepository interface {
	GetTableName() string
	FindByID(ctx context.Context, id uint) (*entity.District, error)
	Find(ctx context.Context, filter *FilterDistrictPayload) ([]*entity.District, int, error)
}

type districtRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewDistrictRepository(db bun.IDB, logger logger.Logger) *districtRepository {
	return &districtRepository{db: db, logger: logger}
}

func (r *districtRepository) GetTableName() string {
	return "districts"
}

type FilterDistrictPayload struct {
	IDs     []uint
	CityIDs []uint
	Names   []string
	Search  string
	Page    int
	PerPage int
}

func (r *districtRepository) Find(ctx context.Context, filter *FilterDistrictPayload) ([]*entity.District, int, error) {
	var districts []*model.District

	query := r.db.NewSelect().Model(&districts)
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.CityIDs) > 0 {
		query = query.Where("city_id IN (?)", bun.In(filter.CityIDs))
	}

	if len(filter.Names) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Names {
				q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Names[i]+"%")
			}

			return q
		})
	}

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Search+"%")
			return q
		})
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count district")
	}

	if totalCount == 0 {
		return []*entity.District{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("id DESC")
	if err := query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find district")
	}

	return model.ToDistrictsDomain(districts), totalCount, nil
}

func (r *districtRepository) FindByID(ctx context.Context, id uint) (*entity.District, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find district by id")
	}

	district := &model.District{ID: id}
	if err := r.db.NewSelect().Model(district).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find district by id")
	}

	return district.ToDomain(), nil
}
