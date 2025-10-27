package mysqlrepository

import (
	"context"
	"database/sql"
	"goapptemp/config"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/pkg/bundb"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ MySQLRepository = (*mysqlRepository)(nil)

type RepositoryAtomicCallback func(r MySQLRepository) error

type MySQLRepository interface {
	DB() *bun.DB
	Atomic(ctx context.Context, config *config.Config, fn RepositoryAtomicCallback) error
	Close() error
	StoreProcedure() StoreProcedureRepository
	Client() ClientRepository
	Role() RoleRepository
	User() UserRepository
	Permission() PermissionRepository
	SupportFeature() SupportFeatureRepository
	Company() CompanyRepository
	Province() ProvinceRepository
	City() CityRepository
	District() DistrictRepository
	ClientSupportFeature() ClientSupportFeatureRepository
}

type mysqlRepository struct {
	db                             bun.IDB
	logger                         logger.Logger
	userRepository                 UserRepository
	companyRepository              CompanyRepository
	clientRepository               ClientRepository
	roleRepository                 RoleRepository
	permissionRepository           PermissionRepository
	supportFeatureRepository       SupportFeatureRepository
	provinceRepository             ProvinceRepository
	districtRepository             DistrictRepository
	cityRepository                 CityRepository
	clientSupportFeatureRepository ClientSupportFeatureRepository
	storeProcedureRepository       StoreProcedureRepository
}

func NewMySQLRepository(config *config.Config, logger logger.Logger) (*mysqlRepository, error) {
	db, err := bundb.NewBunDB(config, logger)
	if err != nil {
		return nil, err
	}

	db.DB().RegisterModel(
		(*model.RolePermission)(nil),
		(*model.UserRole)(nil),
		(*model.City)(nil),
		(*model.Client)(nil),
		(*model.ClientSupportFeature)(nil),
		(*model.Company)(nil),
		(*model.District)(nil),
		(*model.Permission)(nil),
		(*model.Province)(nil),
		(*model.Role)(nil),
		(*model.SupportFeature)(nil),
		(*model.User)(nil),
	)

	return create(config, db.DB(), logger), nil
}

func (r *mysqlRepository) DB() *bun.DB {
	dbInstance, ok := r.db.(*bun.DB)
	if !ok {
		r.logger.Error().Msg("Failed to assert type *bun.DB for the underlying database instance")
		return nil
	}

	return dbInstance
}

func (r *mysqlRepository) Close() error {
	return r.DB().Close()
}

func (r *mysqlRepository) Atomic(ctx context.Context, config *config.Config, fn RepositoryAtomicCallback) error {
	err := r.db.RunInTx(
		ctx,
		&sql.TxOptions{Isolation: sql.LevelSerializable},
		func(ctx context.Context, tx bun.Tx) error {
			return fn(create(config, tx, r.logger))
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func create(config *config.Config, db bun.IDB, logger logger.Logger) *mysqlRepository {
	return &mysqlRepository{
		db:                             db,
		logger:                         logger,
		userRepository:                 NewUserRepository(db, logger),
		clientRepository:               NewClientRepository(db, logger),
		roleRepository:                 NewRoleRepository(db, logger),
		supportFeatureRepository:       NewSupportFeatureRepository(db, logger),
		provinceRepository:             NewProvinceRepository(db, logger),
		cityRepository:                 NewCityRepository(db, logger),
		districtRepository:             NewDistrictRepository(db, logger),
		companyRepository:              NewCompanyRepository(db, logger),
		clientSupportFeatureRepository: NewClientSupportFeatureRepository(db, logger),
		storeProcedureRepository:       NewStoreProcedureRepository(config.MySQL.DBName, db, logger),
		permissionRepository:           NewPermissionRepository(db, logger),
	}
}

func (r *mysqlRepository) User() UserRepository {
	return r.userRepository
}

func (r *mysqlRepository) Company() CompanyRepository {
	return r.companyRepository
}

func (r *mysqlRepository) Client() ClientRepository {
	return r.clientRepository
}

func (r *mysqlRepository) Role() RoleRepository {
	return r.roleRepository
}

func (r *mysqlRepository) Permission() PermissionRepository {
	return r.permissionRepository
}

func (r *mysqlRepository) SupportFeature() SupportFeatureRepository {
	return r.supportFeatureRepository
}

func (r *mysqlRepository) Province() ProvinceRepository {
	return r.provinceRepository
}

func (r *mysqlRepository) City() CityRepository {
	return r.cityRepository
}

func (r *mysqlRepository) District() DistrictRepository {
	return r.districtRepository
}

func (r *mysqlRepository) ClientSupportFeature() ClientSupportFeatureRepository {
	return r.clientSupportFeatureRepository
}

func (r *mysqlRepository) StoreProcedure() StoreProcedureRepository {
	return r.storeProcedureRepository
}
