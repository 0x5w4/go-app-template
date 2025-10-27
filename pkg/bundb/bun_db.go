package bundb

import (
	"context"
	"database/sql"
	"fmt"
	"goapptemp/config"
	"goapptemp/pkg/bundb/hook"
	"goapptemp/pkg/logger"
	"math"
	"time"

	migrationFS "goapptemp/migration"

	"github.com/cockroachdb/errors"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var _ BunDB = (*bunDB)(nil)

type BunDB interface {
	DB() *bun.DB
	Close() error
	Migrate() error
	Reset() error
}

type bunDB struct {
	config *config.Config
	logger logger.Logger
	db     *bun.DB
}

func NewBunDB(config *config.Config, logger logger.Logger) (*bunDB, error) {
	sqlDB, err := sql.Open("mysql", config.MySQL.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}

	// NOTE: This context only used by ping
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.MySQL.ConnMaxLifetime) * time.Minute)

	db := bun.NewDB(sqlDB, mysqldialect.New())
	db.AddQueryHook(hook.NewLoggerHook(
		hook.WithLogger(logger),
		hook.WithDebug(config.MySQL.Debug),
		hook.WithSlowQueryThreshold(time.Duration(config.MySQL.SlowQueryThreshold)*time.Millisecond),
	))
	db.AddQueryHook(hook.NewTracerHook())

	return &bunDB{
		config: config,
		logger: logger,
		db:     db,
	}, nil
}

func (d *bunDB) DB() *bun.DB {
	return d.db
}

func (d *bunDB) Close() error {
	if d.db != nil {
		return d.db.Close()
	}

	return nil
}

func (d *bunDB) Migrate() error {
	sourceInstance, err := iofs.New(migrationFS.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source from embed.FS: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceInstance, d.config.MySQL.MigrateDSN)
	if err != nil {
		return fmt.Errorf("cannot create migration instance: %w", err)
	}

	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			d.logger.Error().Msgf("Error closing migration instance: source_err=%v, db_err=%v", sourceErr, dbErr)
		}
	}()

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	v, err := safeUintToInt(version)
	if err != nil {
		return fmt.Errorf("failed to convert migration version to int: %w", err)
	}

	if dirty {
		d.logger.Warn().Msgf("Dirty migration detected at version %d. Forcing clean state.", version)

		if err := m.Force(v); err != nil {
			return fmt.Errorf("failed to force clean migration state: %w", err)
		}
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, dirty, err = m.Version()
	if err != nil {
		return fmt.Errorf("failed to verify final migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("migration finished in a dirty state at version %d", version)
	}

	d.logger.Info().Msgf("✅ Migration run successfully. Current version: %d", version)

	return nil
}

func (d *bunDB) Reset() error {
	sourceInstance, err := iofs.New(migrationFS.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source from embed.FS: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceInstance, d.config.MySQL.MigrateDSN)
	if err != nil {
		return fmt.Errorf("cannot create migration instance for drop: %w", err)
	}

	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			d.logger.Error().Msgf("Error closing migration instance: source_err=%v, db_err=%v", sourceErr, dbErr)
		}
	}()

	d.logger.Warn().Msg("⚠️ Resetting database by dropping all tables...")

	err = m.Drop()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		m.Close()
		return fmt.Errorf("failed to drop database: %w", err)
	}

	d.logger.Info().Msg("Database reset complete.")
	d.logger.Info().Msg("Applying all migrations...")

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to verify final migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("migration finished in a dirty state at version %d", version)
	}

	d.logger.Info().Msgf("✅ Migration from version 0 completed successfully. Current version: %d", version)

	return nil
}

func safeUintToInt(u uint) (int, error) {
	if u > uint(math.MaxInt) {
		return 0, fmt.Errorf("value %d exceeds max int", u)
	}

	return int(u), nil
}
