package server

import (
	"strings"

	"goapptemp/internal/adapter/api/rest/handler"

	"github.com/labstack/echo/v4"
)

type CRUDHandlers struct {
	Create          echo.HandlerFunc
	BulkCreate      echo.HandlerFunc
	Find            echo.HandlerFunc
	FindOne         echo.HandlerFunc
	Update          echo.HandlerFunc
	Delete          echo.HandlerFunc
	IsDeletable     echo.HandlerFunc
	BulkIsDeletable echo.HandlerFunc
	Export          echo.HandlerFunc
	TemplateImport  echo.HandlerFunc
	ImportPreview   echo.HandlerFunc
}

func (s *server) setupRoutes() {
	s.registerSystemRoutes()
	apiV1 := s.echo.Group("/api/v1")
	s.registerAuthRoutes(apiV1)
	s.registerEntityRoutes(apiV1)
	s.registerWebhookRoutes(apiV1)
}

func (s *server) registerSystemRoutes() {
	healthHandler := handler.NewHealthHandler(s.repo.MySQL().DB(), s.logger)
	s.echo.GET("/ping", healthHandler.CheckHealth)

	if s.config.HTTP.EnableMigrationAPI {
		migrationHandler, err := handler.NewMigrationHandler(s.config, s.logger)
		if err != nil {
			s.logger.Fatal().Err(err).Msg("Failed to create migration handler")
		}

		migrationGroup := s.echo.Group("/sql/migration")
		migrationGroup.GET("/version", migrationHandler.GetVersion)
		migrationGroup.POST("/version", migrationHandler.ForceVersion)
		migrationGroup.GET("/files", migrationHandler.GetMigrationFiles)
		migrationGroup.POST("/up", migrationHandler.Up)
		migrationGroup.POST("/down", migrationHandler.Down)
	}
}

func (s *server) registerAuthRoutes(g *echo.Group) {
	authGroup := g.Group("/auth")
	authGroup.POST("/login", s.handler.Login)
}

func (s *server) registerWebhookRoutes(g *echo.Group) {
	webhookGroup := g.Group("/webhook")
	webhookGroup.POST("/update-icon", s.handler.UpdateIcon)
}

func (s *server) registerEntityRoutes(g *echo.Group) {
	s.registerCRUD(g, "provinces", CRUDHandlers{
		Find:    s.handler.FindProvinces,
		FindOne: s.handler.FindOneProvince,
	}, false)
	s.registerCRUD(g, "cities", CRUDHandlers{
		Find:    s.handler.FindCities,
		FindOne: s.handler.FindOneCity,
	}, false)
	s.registerCRUD(g, "districts", CRUDHandlers{
		Find:    s.handler.FindDistricts,
		FindOne: s.handler.FindOneDistrict,
	}, false)
	s.registerCRUD(g, "users", CRUDHandlers{
		Create:  s.handler.CreateUser,
		Find:    s.handler.FindUsers,
		FindOne: s.handler.FindOneUser,
		Update:  s.handler.UpdateUser,
		Delete:  s.handler.DeleteUser,
	}, true)
	s.registerCRUD(g, "roles", CRUDHandlers{
		Create:  s.handler.CreateRole,
		Find:    s.handler.FindRoles,
		FindOne: s.handler.FindOneRole,
		Update:  s.handler.UpdateRole,
		Delete:  s.handler.DeleteRole,
	}, true)
	s.registerCRUD(g, "help-services", CRUDHandlers{
		Create:         s.handler.CreateSupportFeature,
		BulkCreate:     s.handler.BulkCreateSupportFeatures,
		Find:           s.handler.FindSupportFeatures,
		FindOne:        s.handler.FindOneSupportFeature,
		Update:         s.handler.UpdateSupportFeature,
		Delete:         s.handler.DeleteSupportFeature,
		IsDeletable:    s.handler.IsSupportFeatureDeletable,
		TemplateImport: s.handler.TemplateImportSupportFeature,
		ImportPreview:  s.handler.ImportPreviewSupportFeature,
	}, true)
}

func (s *server) registerCRUD(parentGroup *echo.Group, basePath string, handlers CRUDHandlers, authRequired bool) {
	basePath = strings.TrimPrefix(basePath, "/")

	group := parentGroup.Group("/" + basePath)
	if authRequired {
		group.Use(s.authMiddleware(true))
	}

	if handlers.Create != nil {
		group.POST("", handlers.Create)
	}

	if handlers.BulkCreate != nil {
		group.POST("/bulk", handlers.BulkCreate)
	}

	if handlers.Find != nil {
		group.GET("", handlers.Find)
	}

	if handlers.FindOne != nil {
		group.GET("/:id", handlers.FindOne)
	}

	if handlers.Update != nil {
		group.PUT("/:id", handlers.Update)
	}

	if handlers.Delete != nil {
		group.DELETE("/:id", handlers.Delete)
	}

	if handlers.Export != nil {
		group.GET("/export", handlers.Export)
	}

	if handlers.IsDeletable != nil {
		group.GET("/:id/is-deletable", handlers.IsDeletable)
	}

	if handlers.BulkIsDeletable != nil {
		group.GET("/bulk/is-deletable", handlers.BulkIsDeletable)
	}

	if handlers.TemplateImport != nil {
		group.GET("/template/import", handlers.TemplateImport)
	}

	if handlers.ImportPreview != nil {
		group.POST("/import/preview", handlers.ImportPreview)
	}
}
