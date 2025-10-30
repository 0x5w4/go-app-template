package mysqlrepository

import (
	"context"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"time"

	"github.com/uptrace/bun"
)

var _ CompanyRepository = (*companyRepository)(nil)

type CompanyRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.Company) (*entity.Company, error)
	FindByID(ctx context.Context, id uint) (*entity.Company, error)
	Find(ctx context.Context, filter *FilterCompanyPayload) ([]*entity.Company, int, error)
	Update(ctx context.Context, req *UpdateCompanyPayload) (*entity.Company, error)
	Delete(ctx context.Context, id uint) error
}

type companyRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewCompanyRepository(db bun.IDB, logger logger.Logger) *companyRepository {
	return &companyRepository{db: db, logger: logger}
}

func (r *companyRepository) GetTableName() string {
	return "companies"
}

func (r *companyRepository) Create(ctx context.Context, req *entity.Company) (*entity.Company, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create company")
	}

	company := model.AsCompany(req)
	if _, err := r.db.NewInsert().Model(company).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create company")
	}

	return company.ToDomain(), nil
}

type FilterCompanyPayload struct {
	IDs      []uint
	AdminIDs []uint
	Names    []string
	Search   string
	Page     int
	PerPage  int
}

func (r *companyRepository) Find(ctx context.Context, filter *FilterCompanyPayload) ([]*entity.Company, int, error) {
	var companies []*model.Company

	query := r.db.NewSelect().Model(&companies)
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
		return nil, 0, handleDBError(err, r.GetTableName(), "count company")
	}

	if totalCount == 0 {
		return []*entity.Company{}, 0, nil
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
		return nil, 0, handleDBError(err, r.GetTableName(), "find company")
	}

	return model.ToCompaniesDomain(companies), totalCount, nil
}

func (r *companyRepository) FindByID(ctx context.Context, id uint) (*entity.Company, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find company by id")
	}

	company := &model.Company{Base: model.Base{ID: id}}
	if err := r.db.NewSelect().Model(company).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find company by id")
	}

	return company.ToDomain(), nil
}

type UpdateCompanyPayload struct {
	ID            uint
	Name          *string
	Icon          *string
	IconUpdatedAt *time.Time
	AdminID       *uint
}

func (r *companyRepository) Update(ctx context.Context, req *UpdateCompanyPayload) (*entity.Company, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update company")
	}

	company := &model.Company{Base: model.Base{ID: req.ID}}

	var columnsToUpdate []string

	if req.Name != nil {
		company.Name = *req.Name

		columnsToUpdate = append(columnsToUpdate, "name")
	}

	if req.Icon != nil {
		company.Icon = req.Icon

		columnsToUpdate = append(columnsToUpdate, "icon")
	}

	if req.IconUpdatedAt != nil {
		company.IconUpdatedAt = req.IconUpdatedAt

		columnsToUpdate = append(columnsToUpdate, "icon_updated_at")
	}

	if req.AdminID != nil {
		company.AdminID = *req.AdminID

		columnsToUpdate = append(columnsToUpdate, "admin_id")
	}

	if len(columnsToUpdate) == 0 {
		return company.ToDomain(), nil
	}

	query := r.db.NewUpdate().Model(company).Column(columnsToUpdate...).WherePK().Returning("*")
	if _, err := query.Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update company")
	}

	return company.ToDomain(), nil
}

func (r *companyRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete company")
	}

	company := &model.Company{Base: model.Base{ID: id}}

	_, err := r.db.NewDelete().Model(company).WherePK().Exec(ctx)
	if err != nil {
		return handleDBError(err, r.GetTableName(), "delete company")
	}

	return nil
}
