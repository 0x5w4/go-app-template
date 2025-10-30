package mysqlrepository

import (
	"context"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ ProvinceRepository = (*provinceRepository)(nil)

type ProvinceRepository interface {
	FindByID(ctx context.Context, id uint) (*entity.Province, error)
	Find(ctx context.Context, filter *FilterProvincePayload) ([]*entity.Province, int, error)
}

type provinceRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewProvinceRepository(db bun.IDB, logger logger.Logger) *provinceRepository {
	return &provinceRepository{db: db, logger: logger}
}

func (r *provinceRepository) GetTableName() string {
	return "provinces"
}

type FilterProvincePayload struct {
	IDs     []uint
	Names   []string
	Search  string
	Page    int
	PerPage int
}

func (r *provinceRepository) Find(ctx context.Context, filter *FilterProvincePayload) ([]*entity.Province, int, error) {
	var provinces []*model.Province

	query := r.db.NewSelect().Model(&provinces)
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
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
		return nil, 0, handleDBError(err, r.GetTableName(), "count province")
	}

	if totalCount == 0 {
		return []*entity.Province{}, 0, nil
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
		return nil, 0, handleDBError(err, r.GetTableName(), "find province")
	}

	return model.ToProvincesDomain(provinces), totalCount, nil
}

func (r *provinceRepository) FindByID(ctx context.Context, id uint) (*entity.Province, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find province by id")
	}

	province := &model.Province{ID: id}
	if err := r.db.NewSelect().Model(province).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find province by id")
	}

	return province.ToDomain(), nil
}
