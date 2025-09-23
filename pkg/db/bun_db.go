package db

import (
	"database/sql"
	"fmt"
	"goapptemp/config"
	"goapptemp/pkg/db/hook"
	"goapptemp/pkg/logger"
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

type BunDB struct {
	config *config.Config
	logger logger.Logger
	db     *bun.DB
}

func NewBunDB(config *config.Config, logger logger.Logger) (*BunDB, error) {
	bunDb := &BunDB{
		config: config,
		logger: logger,
	}
	if err := bunDb.connect(); err != nil {
		return nil, err
	}

	return bunDb, nil
}

func (d *BunDB) connect() error {
	sqlDB, err := sql.Open("mysql", d.config.MySQL.DSN)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(d.config.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(d.config.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(d.config.MySQL.ConnMaxLifetime) * time.Minute)

	d.db = bun.NewDB(sqlDB, mysqldialect.New())
	d.db.AddQueryHook(hook.NewLoggerHook(
		hook.WithLogger(d.logger),
		hook.WithDebug(d.config.MySQL.Debug),
		hook.WithSlowQueryThreshold(time.Duration(d.config.MySQL.SlowQueryThreshold)*time.Millisecond),
	))
	d.db.AddQueryHook(hook.NewTracerHook())

	return nil
}

func (d *BunDB) DB() *bun.DB {
	return d.db
}

func (d *BunDB) Close() error {
	if d.db != nil {
		return d.db.Close()
	}

	return nil
}

func (d *BunDB) Migrate() error {
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

	if dirty {
		d.logger.Warn().Msgf("Dirty migration detected at version %d. Forcing clean state.", version)

		if err := m.Force(int(version)); err != nil {
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

func (d *BunDB) Reset() error {
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
