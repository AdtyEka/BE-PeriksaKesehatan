package handler

import (
	"BE-PeriksaKesehatan/config"
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"
	"BE-PeriksaKesehatan/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config, userRepo *repository.UserRepository) *gin.Engine {
	router := gin.Default()

	healthDataRepo := repository.NewHealthDataRepository(userRepo.GetDB())
	authRepo := repository.NewAuthRepository(userRepo.GetDB())
	healthAlertRepo := repository.NewHealthAlertRepository(userRepo.GetDB())
	educationalVideoRepo := repository.NewEducationalVideoRepository(userRepo.GetDB())
	categoryRepo := repository.NewCategoryRepository(userRepo.GetDB())
	healthTargetRepo := repository.NewHealthTargetRepository(userRepo.GetDB())

	healthDataService := service.NewHealthDataService(healthDataRepo)
	healthAlertService := service.NewHealthAlertService(healthAlertRepo, healthDataRepo)
	educationalVideoService := service.NewEducationalVideoService(educationalVideoRepo, categoryRepo)
	profileService := service.NewProfileService(userRepo, healthDataRepo, healthTargetRepo)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(authRepo, cfg.JWTSecret)

	authHandler := NewAuthHandler(userRepo, cfg.JWTSecret)
	healthDataHandler := NewHealthDataHandler(healthDataService, authRepo)
	healthAlertHandler := NewHealthAlertHandler(healthAlertService, authRepo)
	educationalVideoHandler := NewEducationalVideoHandler(educationalVideoService)
	profileHandler := NewProfileHandler(profileService, authRepo)

	api := router.Group("/api")
	{
		// Public routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes (require auth)
		health := api.Group("/health")
		health.Use(authMiddleware)
		{
			health.POST("/data", healthDataHandler.CreateHealthData)
			health.GET("/data", healthDataHandler.GetHealthDataByUserID)
			health.GET("/history", healthDataHandler.GetHealthHistory)
			health.GET("/history/download", healthDataHandler.DownloadHealthReport)
			health.POST("/check-health-alerts", healthAlertHandler.CheckHealthAlerts)
		}

		education := api.Group("/education")
		{
			education.POST("/add-educational-video", educationalVideoHandler.AddEducationalVideo)
			education.GET("/get-educational-videos", educationalVideoHandler.GetAllEducationalVideos)
			education.GET("/get-educational-videos/:id", educationalVideoHandler.GetEducationalVideosByID)
		}

		profile := api.Group("/profile")
		profile.Use(authMiddleware)
		{
			profile.GET("", profileHandler.GetProfile)
			profile.PUT("", profileHandler.UpdateProfile)
			profile.GET("/personal-info", profileHandler.GetPersonalInfo)
			profile.PUT("/personal-info", profileHandler.UpdatePersonalInfo)
			profile.GET("/health-targets", profileHandler.GetHealthTargets)
			profile.PUT("/health-targets", profileHandler.UpdateHealthTargets)
			profile.GET("/settings", profileHandler.GetSettings)
			profile.PUT("/settings", profileHandler.UpdateSettings)
		}
	}

	return router
}

