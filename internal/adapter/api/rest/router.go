package rest

func (s *echoServer) setupRouter() {
	s.echo.GET("/ping", s.handler.Health().CheckHealth)

	if s.config.HTTP.EnableMigrationAPI {
		migrationGroup := s.echo.Group("/sql/migration")
		{
			migrationGroup.GET("/version", s.handler.Migration().GetVersion)
			migrationGroup.POST("/version", s.handler.Migration().ForceVersion)
			migrationGroup.GET("/files", s.handler.Migration().GetMigrationFiles)
			migrationGroup.POST("/up", s.handler.Migration().Up)
			migrationGroup.POST("/down", s.handler.Migration().Down)
		}
	}

	apiV1 := s.echo.Group("/api/v1")
	{
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/login", s.handler.Auth().Login, s.rateLimitMiddleware())
			authGroup.POST("/refresh", s.handler.Auth().Refresh)
			authGroup.POST("/logout", s.handler.Auth().Logout)
			authGroup.POST("/forget-password", s.handler.Auth().ForgetPassword, s.rateLimitMiddleware())
			authGroup.POST("/verify-reset-token", s.handler.Auth().VerifyResetToken)
			authGroup.POST("/reset-password", s.handler.Auth().ResetPassword)
		}

		webhookGroup := apiV1.Group("/webhook")
		{
			webhookGroup.POST("/update-icon", s.handler.Webhook().UpdateIcon)
		}

		provinceGroup := apiV1.Group("/provinces")
		provinceGroup.Use(s.authMiddleware(false))
		{
			provinceGroup.GET("", s.handler.Province().FindProvinces)
			provinceGroup.GET("/:id", s.handler.Province().FindOneProvince)
		}

		cityGroup := apiV1.Group("/cities")
		cityGroup.Use(s.authMiddleware(false))
		{
			cityGroup.GET("", s.handler.City().FindCities)
			cityGroup.GET("/:id", s.handler.City().FindOneCity)
		}

		districtGroup := apiV1.Group("/districts")
		districtGroup.Use(s.authMiddleware(false))
		{
			districtGroup.GET("", s.handler.District().FindDistricts)
			districtGroup.GET("/:id", s.handler.District().FindOneDistrict)
		}

		userGroup := apiV1.Group("/users")
		userGroup.Use(s.authMiddleware(false))
		{
			userGroup.GET("", s.handler.User().FindUsers)
			userGroup.GET("/:id", s.handler.User().FindOneUser)
			userGroup.POST("", s.handler.User().CreateUser, s.authMiddleware(true))
			userGroup.PUT("/:id", s.handler.User().UpdateUser, s.authMiddleware(true))
			userGroup.DELETE("/:id", s.handler.User().DeleteUser, s.authMiddleware(true))
		}

		roleGroup := apiV1.Group("/roles")
		roleGroup.Use(s.authMiddleware(false))
		{
			roleGroup.POST("", s.handler.Role().CreateRole)
			roleGroup.GET("", s.handler.Role().FindRoles)
			roleGroup.GET("/:id", s.handler.Role().FindOneRole)
			roleGroup.PUT("/:id", s.handler.Role().UpdateRole)
			roleGroup.DELETE("/:id", s.handler.Role().DeleteRole)
		}

		supportFeatureGroup := apiV1.Group("/help-services")
		supportFeatureGroup.Use(s.authMiddleware(false))
		{
			supportFeatureGroup.POST("", s.handler.SupportFeature().CreateSupportFeature)
			supportFeatureGroup.POST("/bulk", s.handler.SupportFeature().BulkCreateSupportFeatures)
			supportFeatureGroup.GET("", s.handler.SupportFeature().FindSupportFeatures)
			supportFeatureGroup.GET("/:id", s.handler.SupportFeature().FindOneSupportFeature)
			supportFeatureGroup.PUT("/:id", s.handler.SupportFeature().UpdateSupportFeature)
			supportFeatureGroup.DELETE("/:id", s.handler.SupportFeature().DeleteSupportFeature)
			supportFeatureGroup.GET("/:id/is-deletable", s.handler.SupportFeature().IsSupportFeatureDeletable)
			supportFeatureGroup.GET("/template/import", s.handler.SupportFeature().TemplateImportSupportFeature)
			supportFeatureGroup.POST("/import/preview", s.handler.SupportFeature().ImportPreviewSupportFeature)
		}
	}
}
