package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	migrate "github.com/golang-migrate/migrate/v4"
	echo "github.com/labstack/echo/v4"
)

type MigrationHandler struct {
	properties
}

func NewMigrationHandler(properties properties) *MigrationHandler {
	return &MigrationHandler{
		properties: properties,
	}
}

func (h *MigrationHandler) GetVersion(c echo.Context) error {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration path"})
	}

	m, err := migrate.New(migrationPath, h.config.MySQL.MigrateDSN)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create migration instance")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create migration instance"})
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			h.logger.Error().Err(srcErr).Err(dbErr).Msg("Failed to close migration instance")
		}
	}()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return c.JSON(http.StatusOK, echo.Map{
				"version": "No migrations applied yet",
				"dirty":   false,
			})
		}

		h.logger.Error().Err(err).Msg("Failed to get migration version")

		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration version"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"version": version,
		"dirty":   dirty,
	})
}

func (h *MigrationHandler) ForceVersion(c echo.Context) error {
	versionStr := c.QueryParam("version")
	if versionStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "version parameter is required"})
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid version parameter"})
	}

	migrationPath, err := getMigrationPath()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration path"})
	}

	m, err := migrate.New(migrationPath, h.config.MySQL.MigrateDSN)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create migration instance")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create migration instance"})
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			h.logger.Error().Err(srcErr).Err(dbErr).Msg("Failed to close migration instance")
		}
	}()

	if err := m.Force(version); err != nil {
		h.logger.Error().Err(err).Msgf("Failed to force migration version to %d", version)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": fmt.Sprintf("failed to force migration version to %d", version)})
	}

	h.logger.Info().Msgf("Successfully forced migration version to %d", version)

	return c.JSON(http.StatusOK, echo.Map{"message": fmt.Sprintf("successfully forced migration version to %d", version)})
}

func (h *MigrationHandler) GetMigrationFiles(c echo.Context) error {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration path"})
	}

	pathOnly := migrationPath[len("file://"):]

	files, err := os.ReadDir(pathOnly)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read migration directory")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to read migration directory"})
	}

	var fileNames []string

	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"files": fileNames})
}

func (h *MigrationHandler) Up(c echo.Context) error {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration path"})
	}

	m, err := migrate.New(migrationPath, h.config.MySQL.MigrateDSN)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create migration instance")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create migration instance"})
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			h.logger.Error().Err(srcErr).Err(dbErr).Msg("Failed to close migration instance")
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return c.JSON(http.StatusOK, echo.Map{"message": "no new migrations to apply"})
		}

		h.logger.Error().Err(err).Msg("Migration up failed")

		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "migration up failed"})
	}

	h.logger.Info().Msg("Migrations applied successfully")

	return c.JSON(http.StatusOK, echo.Map{"message": "migrations applied successfully"})
}

func (h *MigrationHandler) Down(c echo.Context) error {
	migrationPath, err := getMigrationPath()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get migration path"})
	}

	m, err := migrate.New(migrationPath, h.config.MySQL.MigrateDSN)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create migration instance")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create migration instance"})
	}

	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			h.logger.Error().Err(srcErr).Err(dbErr).Msg("Failed to close migration instance")
		}
	}()

	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return c.JSON(http.StatusOK, echo.Map{"message": "no migrations to revert"})
		}

		h.logger.Error().Err(err).Msg("Migration down failed")

		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "migration down failed"})
	}

	h.logger.Info().Msg("Migration reverted successfully")

	return c.JSON(http.StatusOK, echo.Map{"message": "migration reverted successfully"})
}

func getMigrationPath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("unable to get current file path")
	}

	handlerDir := filepath.Dir(currentFile)
	dbDir := filepath.Join(handlerDir, "..", "..", "..", "repository", "mysql", "db")

	migrationPath := filepath.ToSlash(path.Join(dbDir, "migration"))
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return "", fmt.Errorf("migration directory not found at: %s", migrationPath)
	}

	return "file://" + migrationPath, nil
}
