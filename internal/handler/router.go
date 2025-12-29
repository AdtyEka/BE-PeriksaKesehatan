package handler

import (
	"BE-PeriksaKesehatan/internal/repository"
	"BE-PeriksaKesehatan/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter mengkonfigurasi semua route aplikasi
func SetupRouter(userRepo *repository.UserRepository) *gin.Engine {
	// Setup Gin dengan default middleware
	router := gin.Default()

	// Setup repository
	healthDataRepo := repository.NewHealthDataRepository(userRepo.GetDB())
	authRepo := repository.NewAuthRepository(userRepo.GetDB())
	healthAlertRepo := repository.NewHealthAlertRepository(userRepo.GetDB())
	educationalVideoRepo := repository.NewEducationalVideoRepository(userRepo.GetDB())

	// Setup service
	healthDataService := service.NewHealthDataService(healthDataRepo)
	healthAlertService := service.NewHealthAlertService(healthAlertRepo, healthDataRepo)
	educationalVideoService := service.NewEducationalVideoService(educationalVideoRepo)

	// Setup handler
	authHandler := NewAuthHandler(userRepo)
	healthDataHandler := NewHealthDataHandler(healthDataService, authRepo)
	healthAlertHandler := NewHealthAlertHandler(healthAlertService, authRepo)
	educationalVideoHandler := NewEducationalVideoHandler(educationalVideoService)

	// API Routes
	api := router.Group("/api")
	{
		// Route autentikasi (register, login & logout)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Route data kesehatan (memerlukan autentikasi)
		health := api.Group("/health")
		{
			health.POST("/data", healthDataHandler.CreateHealthData)
			health.GET("/data", healthDataHandler.GetHealthDataByUserID)
			health.GET("/history", healthDataHandler.GetHealthHistory)                    // Endpoint riwayat kesehatan
			health.GET("/history/download", healthDataHandler.DownloadHealthReport)       // Endpoint download laporan
			health.POST("/check-health-alerts", healthAlertHandler.CheckHealthAlerts)     // Endpoint check health alerts
		}

		// Route video edukasi kesehatan
		education := api.Group("/education")
		{
			education.POST("/add-educational-video", educationalVideoHandler.AddEducationalVideo)              // Endpoint tambah video
			education.GET("/get-educational-videos/:health_condition", educationalVideoHandler.GetEducationalVideos) // Endpoint ambil video
		}
	}

	return router
}

