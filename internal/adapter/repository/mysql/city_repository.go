package mysqlrepository

import (
	"context"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ CityRepository = (*cityRepository)(nil)

type CityRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.City, error)
	Find(ctx context.Context, filter *FilterCityPayload) ([]*entity.City, int, error)
}

type cityRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewCityRepository(db bun.IDB, logger logger.Logger) *cityRepository {
	return &cityRepository{db: db, logger: logger}
}

func (r *cityRepository) GetTableName() string {
	return "cities"
}

type FilterCityPayload struct {
	IDs         []uint
	ProvinceIDs []uint
	Names       []string
	Search      string
	Page        int
	PerPage     int
}

func (r *cityRepository) Find(ctx context.Context, filter *FilterCityPayload) ([]*entity.City, int, error) {
	var cities []*model.City

	query := r.db.NewSelect().Model(&cities)
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.ProvinceIDs) > 0 {
		query = query.Where("province_id IN (?)", bun.In(filter.ProvinceIDs))
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
		return nil, 0, handleDBError(err, r.GetTableName(), "count city")
	}

	if totalCount == 0 {
		return []*entity.City{}, 0, nil
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
		return nil, 0, handleDBError(err, r.GetTableName(), "find city")
	}

	return model.ToCitiesDomain(cities), totalCount, nil
}

func (r *cityRepository) FindByID(ctx context.Context, id uint) (*entity.City, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find city by id")
	}

	city := &model.City{ID: id}
	if err := r.db.NewSelect().Model(city).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find city by id")
	}

	return city.ToDomain(), nil
}
